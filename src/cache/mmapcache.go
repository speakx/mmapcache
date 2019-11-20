package cache

import (
	"errors"
	"os"

	"github.com/edsrzf/mmap-go"

	"mmapcache/byteio"
)

const (
	mmapCacheHeadSize        = 1024
	mmapCacheHeadVersionPos  = 4
	mmapCacheHeadStatusPos   = mmapCacheHeadVersionPos + 2
	mmapCacheHeadDataSizePos = mmapCacheHeadStatusPos + 2
	mmapCacheContentPos      = mmapCacheHeadSize
)

// MMapCache 基于mmap模式的文件缓存
// | ---------------------- head ---------------------------------------| --- content ---|
// | 4byte:content.len | 2byte:version  | 2byte:status | 4byte:datasize |   mmapdata.go  |
type MMapCache struct {
	path             string
	f                *os.File
	buf              []byte // mmap后的文件原始内存
	writeContent     []byte // content部分的内存对象
	writeUint32Cache []byte // 内存写缓存，保证一次copy写内存，防止按字节写出错
	dataSize         int    // 每次data的固定分配长度，可以支持快速写，但是弊端就是需要提前设计好将要写入的数据最大长度，否则会有数据写失败
	readPos          int
	writePos         int
	mmapdataIdx      map[string]*MMapData
	mmapdataAry      []*MMapData
}

func newMMapCache(filePath string, dataSize int, reload bool) (*MMapCache, error) {
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0)
	if nil != err {
		return nil, err
	}
	buf, err := mmap.MapRegion(f, -1, mmap.RDWR, 0, 0)
	if nil != err {
		return nil, err
	}

	mmcache := &MMapCache{
		path:             filePath,
		f:                f,
		buf:              buf,
		writeContent:     buf[mmapCacheContentPos:],
		writeUint32Cache: make([]byte, 4),
		dataSize:         dataSize,
	}

	byteio.Uint16ToBytes(uint16(0x1), mmcache.buf[mmapCacheHeadVersionPos:])
	byteio.Uint32ToBytes(uint32(dataSize), mmcache.buf[mmapCacheHeadDataSizePos:])
	mmcache.init(reload)
	return mmcache, nil
}

// ReloadMMapCache 通过内存对象
// 反序列化出之前的MMapCache对象与MMapCache对象中的MMapData
func ReloadMMapCache(buf []byte) *MMapCache {
	mmcache := &MMapCache{
		buf:          buf,
		writeContent: buf[mmapCacheContentPos:],
	}

	mmcache.init(true)
	return mmcache
}

// Release 释放，将此mmap文件丢到pool中，由pool的策略决定释放真正释放
func (m *MMapCache) Release() {
	if nil != m.f {
		DefPoolMMapCache.Collect(m)
	}
}

// Path mmap文件所映射的本地文件对象
func (m *MMapCache) Path() string {
	return m.path
}

// GetMMapDatas 获取当前Cache文件中存储的所有mmapdata对象
// 当通过Reload加载完毕MMapCache文件后，调用此方法获取到所有文件内的对象数据，然后通过反序列化初始化出内存对象
// for _, mmapdata := range GetMMapDatas() {
//     val := ...Unmarshal(mmapdata.GetData())
//     mmapdata.ReloadVal(val)
// }
func (m *MMapCache) GetMMapDatas() []*MMapData {
	return m.mmapdataAry
}

// WriteData 写入一片内存对象
// 返回 (-1, nil) 表示当前mmap对象已无可用空间
// 返回 (0, error)，表示当前的待写入对象，超出了mmap对象的datasize
func (m *MMapCache) WriteData(tag uint16, data, key []byte, val interface{}) (int, error) {
	// 判断是否已经有这个缓存了
	mmapData, _ := m.mmapdataIdx[string(key)]
	if nil == mmapData {
		if m.writePos+m.dataSize+mmapCacheHeadSize > len(m.buf) {
			return -1, nil
		}

		writeBuf := m.writeContent[m.writePos:]
		mmapData = newMMapData(uint32(m.dataSize), tag, writeBuf, key, data, val)
		if nil == mmapData {
			return 0, errors.New("mmap cache data.size over follow")
		}

		m.mmapdataIdx[string(key)] = mmapData
		m.mmapdataAry = append(m.mmapdataAry, mmapData)
		m.setWritePos(m.writePos + m.dataSize)
	}

	mmapData.writeData(data)
	return len(data), nil
}

// GetWrittenData 返回有数据的mmap内存
func (m *MMapCache) GetWrittenData() []byte {
	return m.buf[:mmapCacheHeadSize+m.writePos]
}

// GetFreeContentLen 返回可写入的空余内存大小
func (m *MMapCache) GetFreeContentLen() int {
	return len(m.writeContent) - m.writePos
}

// SetStatus 设置自定义状态
func (m *MMapCache) SetStatus(s uint16) {
	byteio.Uint16ToBytes(s, m.buf[mmapCacheHeadStatusPos:])
}

// GetStatus 获取自定义状态
func (m *MMapCache) GetStatus() uint16 {
	return byteio.BytesToUint16(m.buf[mmapCacheHeadStatusPos:])
}

func (m *MMapCache) close(remove bool) {
	m.f.Close()
	if remove {
		os.Remove(m.path)
	}
}

func (m *MMapCache) setWritePos(n int) {
	byteio.SafeUint32ToBytes(uint32(n), m.buf, m.writeUint32Cache)
	m.writePos = n
}

func (m *MMapCache) getWritePos() int {
	return int(byteio.BytesToUint32(m.buf))
}

func (m *MMapCache) init(reload bool) {

	if reload {
		m.readPos = 0
		m.writePos = m.getWritePos()
		m.dataSize = int(byteio.BytesToUint32(m.buf[mmapCacheHeadDataSizePos:]))

		m.mmapdataIdx = make(map[string]*MMapData)
		m.mmapdataAry = make([]*MMapData, 0, (len(m.buf)-mmapCacheHeadSize)/m.dataSize)

		reloadBuf := m.writeContent
		for i := m.writePos; i > 0; {
			mmapData := reloadMMapData(reloadBuf)
			i -= int(mmapData.GetSize())
			m.mmapdataAry = append(m.mmapdataAry, mmapData)
			reloadBuf = reloadBuf[mmapData.GetSize():]
			m.mmapdataIdx[string(mmapData.GetKey())] = mmapData
		}
	} else {
		m.readPos = 0
		m.setWritePos(0)

		m.mmapdataIdx = make(map[string]*MMapData)
		m.mmapdataAry = make([]*MMapData, 0, (len(m.buf)-mmapCacheHeadSize)/m.dataSize)
	}
}

func (m *MMapCache) recycle(template []byte) {
	m.init(false)
}

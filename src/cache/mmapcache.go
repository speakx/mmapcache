package cache

import (
	"errors"
	"os"

	"github.com/edsrzf/mmap-go"

	"mmapcache/byteio"
)

const (
	mmapCacheHeadSize        = 1024
	mmapCacheHeadIDPos       = 4
	mmapCacheHeadNextIDPos   = mmapCacheHeadIDPos + 8
	mmapCacheHeadVersionPos  = mmapCacheHeadNextIDPos + 8
	mmapCacheHeadDataSizePos = mmapCacheHeadVersionPos + 8
	mmapCacheContentPos      = mmapCacheHeadSize
)

// MMapCache 基于mmap模式的文件缓存
// | ------------------- head ----------------------------------------------------| --- content ---|
// | 4byte:content.len | 8byte:id | 8byte:nextid | 2byte:version | 4byte:datasize |   mmapdata.go  |
type MMapCache struct {
	path             string
	f                *os.File
	buf              []byte // mmap后的文件原始内存
	writeContent     []byte // content部分的内存对象
	writeUint32Cache []byte // 内存写缓存，保证一次copy写内存，防止按字节写出错
	dataSize         int    // 每次data的固定分配长度，可以支持快速写，但是弊端就是需要提前设计好将要写入的数据最大长度，否则会有数据写失败
	readPos          int
	writePos         int
	nextMMapCache    *MMapCache
	mmapdataIdx      map[string]*MMapData
	mmapdataAry      []*MMapData
}

func newMMapCache(filePath string, dataSize int) (*MMapCache, error) {
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

	mmcache.init(true)
	return mmcache, nil
}

// GetID 获取当前mmap对象的ID
func (m *MMapCache) GetID() uint64 {
	return byteio.BytesToUint64(m.getIDBuf())
}

// GetNextID 当多片mmap合并到一起使用时获取到下一片的mmap对象
func (m *MMapCache) GetNextID() uint64 {
	return byteio.BytesToUint64(m.getNextIDBuf())
}

// MergeMMapCache 将另一片mmap与当前mmap合并到一起产生一个新的大块mmap对象
func (m *MMapCache) MergeMMapCache(mmapCache *MMapCache) *MMapCache {
	mmapCache.setNext(m)
	return mmapCache
}

// Release 释放，将此mmap文件丢到pool中，由pool的策略决定释放真正释放
func (m *MMapCache) Release() {
	mmapCache := m
	for {
		next := mmapCache.nextMMapCache
		DefPoolMMapCache.Collect(mmapCache)

		if nil == next {
			break
		}
	}
}

// Write 写入一片内存对象
// 返回 -1 表示当前mmap对象已无可用空间
// 返回 error，表示当前的待写入对象，超出了mmap对象的datasize
func (m *MMapCache) Write(tag uint16, data, key []byte, val interface{}) (int, error) {
	if len(data)+mmapDataHeadLen > m.dataSize {
		return 0, errors.New("mmap cache data.size over follow")
	}

	// 判断是否已经有这个缓存了
	mmapData, _ := m.mmapdataIdx[string(key)]
	if nil == mmapData {
		if m.writePos+m.dataSize+mmapCacheHeadSize > len(m.buf) {
			return -1, nil
		}

		writeBuf := m.writeContent[m.writePos:]
		mmapData = newMMapData(uint32(m.dataSize), tag, writeBuf, val)
		m.mmapdataIdx[string(key)] = mmapData
		m.mmapdataAry = append(m.mmapdataAry, mmapData)
		m.setWritePos(m.writePos + m.dataSize)
	}

	mmapData.writeData(data)
	return len(data), nil
}

// Read 按照写入的顺序，顺序读出byte数据块
// 如果p为nil，则返回当前读取一次，需要的byte.len
// 如果p为nil，且返回的n也为0，则表示没有可读取数据
func (m *MMapCache) Read(p []byte) (int, error) {
	return 0, nil
}

// GetFreeDataLen 获取剩下的可以write的数据空间
func (m *MMapCache) GetFreeDataLen() int {
	return len(m.writeContent) - m.writePos
}

func (m *MMapCache) name(id uint64) {
	byteio.Uint64ToBytes(id, m.getIDBuf())
}

func (m *MMapCache) close(remove bool) {
	m.f.Close()
	if remove {
		os.Remove(m.path)
	}
}

func (m *MMapCache) getIDBuf() []byte {
	return m.buf[mmapCacheHeadIDPos:]
}

func (m *MMapCache) getNextIDBuf() []byte {
	return m.buf[mmapCacheHeadNextIDPos:]
}

func (m *MMapCache) setNext(mmapCache *MMapCache) {
	if nil != mmapCache {
		byteio.Uint64ToBytes(mmapCache.GetID(), m.getNextIDBuf())
		m.nextMMapCache = mmapCache
	} else {
		byteio.Uint64ToBytes(0, m.getNextIDBuf())
		m.nextMMapCache = nil
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
		// m.dataSize = int(byteio.BytesToUint32(m.buf[mmapCacheHeadDataSizePos:]))
		m.readPos = 0
		m.writePos = m.getWritePos()
	} else {
		m.readPos = 0
		m.setWritePos(0)
		m.setNext(nil)
	}
	m.mmapdataIdx = make(map[string]*MMapData)
	m.mmapdataAry = make([]*MMapData, 0, (len(m.buf)-mmapCacheHeadSize)/m.dataSize)
}

func (m *MMapCache) recycle(template []byte) {
	m.init(false)
}

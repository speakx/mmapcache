package cache

import (
	"mmapcache/byteio"
)

const (
	mmapDataHeadLen       = 12
	mmapDataHeadUsedPos   = 4
	mmapDataHeadTagPos    = mmapDataHeadUsedPos + 4
	mmapDataHeadKeyLenPos = mmapDataHeadTagPos + 2
	mmapDataPos           = mmapDataHeadLen
)

// MMapData mmap数据块
// | ------------------------------------------- head ----------------------------------------| ---------- data ---------- |
// | -- 4byte:data.size -- | -- 4byte:data.used -- | -- 2byte:datatag -- | -- 2byte:keylen -- | -- keydata -- | -- data -- |
type MMapData struct {
	buf     []byte
	data    []byte
	keyLen  int
	dataLen int
	val     interface{}
}

// GetSize 返回Size
func (m *MMapData) GetSize() uint32 {
	return byteio.BytesToUint32(m.buf)
}

// GetTag 返回Tag
func (m *MMapData) GetTag() uint16 {
	return byteio.BytesToUint16(m.buf[mmapDataHeadTagPos:])
}

// GetKey 返回Key
func (m *MMapData) GetKey() []byte {
	return m.buf[mmapDataPos : mmapDataPos+m.keyLen]
}

// GetData 返回Data
func (m *MMapData) GetData() []byte {
	return m.data[:m.dataLen]
}

// GetVal 返回Val
func (m *MMapData) GetVal() interface{} {
	return m.val
}

// ReloadVal 通过GetData方法可获取之前val对象序列化后的数据
// 反序列化后可继续通过Val属性绑定文件缓存与内存对象之间的关系
func (m *MMapData) ReloadVal(val interface{}) {
	m.val = val
}

func reloadMMapData(buf []byte) *MMapData {
	mmapData := &MMapData{
		buf: buf,
	}
	mmapData.keyLen = int(byteio.BytesToUint16(buf[mmapDataHeadKeyLenPos:]))
	mmapData.dataLen = int(byteio.BytesToUint32(buf[mmapDataHeadUsedPos:])) - mmapData.keyLen
	mmapData.data = mmapData.buf[mmapDataHeadLen+mmapData.keyLen:]
	return mmapData
}

func newMMapData(size uint32, tag uint16, buf, key, data []byte, val interface{}) *MMapData {
	used := len(key) + len(data)
	if (used + mmapDataHeadLen) > int(size) {
		return nil
	}

	mmapData := &MMapData{
		buf:     buf,
		data:    buf[mmapDataHeadLen+len(key):],
		keyLen:  len(key),
		dataLen: len(data),
		val:     val,
	}
	byteio.Uint32ToBytes(size, mmapData.buf)
	byteio.Uint16ToBytes(tag, mmapData.buf[mmapDataHeadTagPos:])
	byteio.Uint16ToBytes(uint16(len(key)), mmapData.buf[mmapDataHeadKeyLenPos:])
	copy(mmapData.buf[mmapDataPos:], key)

	mmapData.writeData(data)
	return mmapData
}

func (m *MMapData) writeData(data []byte) {
	m.dataLen = len(data)
	byteio.Uint32ToBytes(uint32(m.keyLen+m.dataLen), m.buf[mmapDataHeadUsedPos:])
	copy(m.data, data)
}

func (m *MMapData) getUsed() uint32 {
	return byteio.BytesToUint32(m.buf[mmapDataHeadUsedPos:])
}

func (m *MMapData) getKeyUsed() uint32 {
	return uint32(byteio.BytesToUint16(m.buf[mmapDataHeadKeyLenPos:]))
}

func (m *MMapData) getDataUsed() uint32 {
	return byteio.BytesToUint32(m.buf[mmapDataHeadUsedPos:]) - m.getKeyUsed()
}

func (m *MMapData) getHead() []byte {
	return m.buf[:mmapDataPos]
}

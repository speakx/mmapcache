package cache

import (
	"mmapcache/byteio"
)

const (
	mmapDataHeadLen     = 10
	mmapDataHeadUsedPos = 4
	mmapDataHeadTagPos  = mmapDataHeadUsedPos + 4
	mmapDataPos         = mmapDataHeadLen
)

// MMapData mmap数据块
// | -- 4byte:data.size -- | -- 4byte:data.used -- | -- 2byte:datatag -- | -- data -- |
type MMapData struct {
	buf []byte
	val interface{}
}

func newMMapData(size uint32, tag uint16, buf []byte, val interface{}) *MMapData {
	mmapData := &MMapData{
		buf: buf,
		val: val,
	}
	byteio.Uint32ToBytes(size, mmapData.buf)
	byteio.Uint16ToBytes(tag, mmapData.buf[mmapDataHeadTagPos:])
	return mmapData
}

func (m *MMapData) writeData(data []byte) {
	byteio.Uint32ToBytes(uint32(len(data)), m.buf[mmapDataHeadUsedPos:])
	copy(m.buf[mmapDataPos:], data)
}

func (m *MMapData) getUsed() uint32 {
	return byteio.BytesToUint32(m.buf[mmapDataHeadUsedPos:])
}

func (m *MMapData) getTag() uint16 {
	return byteio.BytesToUint16(m.buf[mmapDataHeadTagPos:])
}

func (m *MMapData) getData() []byte {
	return m.buf[mmapDataPos : mmapDataPos+m.getUsed()]
}

func (m *MMapData) getHead() []byte {
	return m.buf[:mmapDataPos]
}

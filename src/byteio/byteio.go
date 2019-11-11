package byteio

// Uint64ToBytes 将uint64写入到byte中（本机字节序）
// 注意：本方法非原子函数，如果写入过程中崩溃会导致数据出错（SafeUint64ToBytes为原子函数）
func Uint64ToBytes(n uint64, buf []byte) {
	buf[7] = uint8(n)
	buf[6] = uint8(n >> 8)
	buf[5] = uint8(n >> 16)
	buf[4] = uint8(n >> 24)
	buf[3] = uint8(n >> 32)
	buf[2] = uint8(n >> 40)
	buf[1] = uint8(n >> 48)
	buf[0] = uint8(n >> 56)
}

// SafeUint64ToBytes 将uint64写入到byte中（本机字节序）
func SafeUint64ToBytes(n uint64, buf, cache []byte) {
	cache[7] = uint8(n)
	cache[6] = uint8(n >> 8)
	cache[5] = uint8(n >> 16)
	cache[4] = uint8(n >> 24)
	cache[3] = uint8(n >> 32)
	cache[2] = uint8(n >> 40)
	cache[1] = uint8(n >> 48)
	cache[0] = uint8(n >> 56)
	copy(buf, cache)
}

// BytesToUint64 将byte中的8字节数据，转换为uint64（本机字节序）
func BytesToUint64(buf []byte) uint64 {
	return uint64(buf[0])<<56 | uint64(buf[1])<<48 | uint64(buf[2])<<40 | uint64(buf[3])<<32 | uint64(buf[4])<<24 | uint64(buf[5])<<16 | uint64(buf[6])<<8 | uint64(buf[7])
}

// Uint32ToBytes 将uint32写入到byte中（本机字节序）
// 注意：本方法非原子函数，如果写入过程中崩溃会导致数据出错（SafeUint32ToBytes为原子函数）
func Uint32ToBytes(n uint32, buf []byte) {
	buf[3] = uint8(n)
	buf[2] = uint8(n >> 8)
	buf[1] = uint8(n >> 16)
	buf[0] = uint8(n >> 24)
}

// SafeUint32ToBytes 将uint32写入到byte中（本机字节序）
func SafeUint32ToBytes(n uint32, buf, cache []byte) {
	cache[3] = uint8(n)
	cache[2] = uint8(n >> 8)
	cache[1] = uint8(n >> 16)
	cache[0] = uint8(n >> 24)
	copy(buf, cache)
}

// BytesToUint32 将byte中的4字节数据，转换为uint32（本机字节序）
func BytesToUint32(buf []byte) uint32 {
	return uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])
}

// Uint16ToBytes 将uint16写入到byte中（本机字节序）
// 注意：本方法非原子函数，如果写入过程中崩溃会导致数据出错
func Uint16ToBytes(n uint16, buf []byte) {
	buf[1] = uint8(n)
	buf[0] = uint8(n >> 8)
}

// BytesToUint16 将byte中的2字节数据，转换为uint16（本机字节序）
func BytesToUint16(buf []byte) uint16 {
	return uint16(buf[0])<<8 | uint16(buf[1])
}

// Uint8ToBytes 将uint8写入到byte中（无关字节序、原子操作）
func Uint8ToBytes(n uint8, buf []byte) {
	buf[0] = n
}

// BytesToUint8 将byte中的1字节数据，转换为uint8
func BytesToUint8(buf []byte) uint8 {
	return uint8(buf[0])
}

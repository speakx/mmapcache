package byteio

import (
	"bytes"
	"encoding/binary"
	"testing"
)

var buf []byte
var cache []byte

var u16 uint16
var u16Read uint16

func TestUint16ToBytes(t *testing.T) {
	buf = make([]byte, 10)
	u16 = 0xABCD
	t.Logf("uint16:%x %v", u16, u16)
	t.Logf("buf:%v", buf)

	Uint16ToBytes(u16, buf)
	t.Logf("Uint16ToBytes buf:%v", buf)

	u16Read = BytesToUint16(buf)
	t.Logf("u16Read:%x %v", u16Read, u16Read)
	if u16Read != u16 {
		t.Errorf("Uint16ToBytes != BytesToUint16 buf:%v", buf)
	}
}

var u32 uint32
var u32Read uint32

func TestUint32ToBytes(t *testing.T) {
	buf = make([]byte, 8)
	u32 = uint32(0xabcdef12)
	t.Logf("uint32:%x %v", u32, u32)
	t.Logf("0xab:%v 0xcd:%v 0xef:%v 0x12:%v", 0xab, 0xcd, 0xef, 0x12)
	t.Logf("buf:%v", buf)

	Uint32ToBytes(u32, buf)
	t.Logf("Uint32ToBytes buf:%v", buf)

	buf = make([]byte, 8)
	cache = make([]byte, 4)
	t.Logf("realloc buf:%v", buf)
	SafeUint32ToBytes(u32, buf, cache)
	t.Logf("SafeUint32ToBytes buf:%v", buf)

	u32Read = BytesToUint32(buf)
	t.Logf("u32Read:%x %v", u32Read, u32Read)
	if u32Read != u32 {
		t.Errorf("Uint32ToBytes != BytesToUint32 buf:%v", buf)
	}
}

func Benchmark_Uint32ToBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Uint32ToBytes(u32, buf)
	}
}

func Benchmark_Uint32ToBytes_Bytes(b *testing.B) {
	buffer := bytes.NewBuffer(buf)
	for i := 0; i < b.N; i++ {
		binary.Write(buffer, binary.BigEndian, u32)
	}
}

func Benchmark_BytesToUint32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BytesToUint32(buf)
	}
}

func Benchmark_BytesToUint32_Bytes(b *testing.B) {
	buffer := bytes.NewBuffer(buf)
	for i := 0; i < b.N; i++ {
		binary.Read(buffer, binary.BigEndian, &u32Read)
	}
}

var u64 uint64
var u64Read uint64

func TestUint64ToBytes(t *testing.T) {
	buf = make([]byte, 20)
	u64 = 0xABCDEFABCDEF1234
	t.Logf("uint64:%x %v", u64, u64)
	t.Logf("buf:%v", buf)

	Uint64ToBytes(u64, buf)
	t.Logf("Uint64ToBytes buf:%v", buf)

	buf = make([]byte, 20)
	cache = make([]byte, 8)
	t.Logf("realloc buf:%v", buf)
	SafeUint64ToBytes(u64, buf, cache)
	t.Logf("SafeUint64ToBytes buf:%v", buf)

	u64Read = BytesToUint64(buf)
	t.Logf("u64Read:%x %v", u64Read, u64Read)
	if u64Read != u64 {
		t.Errorf("Uint64ToBytes != BytesToUint64 buf:%v", buf)
	}
}

func Benchmark_Uint64ToBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Uint64ToBytes(u64, buf)
	}
}

func Benchmark_Uint64ToBytes_Bytes(b *testing.B) {
	buffer := bytes.NewBuffer(buf)
	for i := 0; i < b.N; i++ {
		binary.Write(buffer, binary.BigEndian, u64)
	}
}

func Benchmark_BytesToUint64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BytesToUint64(buf)
	}
}

func Benchmark_BytesToUint64_Bytes(b *testing.B) {
	buffer := bytes.NewBuffer(buf)
	for i := 0; i < b.N; i++ {
		binary.Read(buffer, binary.BigEndian, &u64Read)
	}
}

package byteio

import (
	"reflect"
	"testing"
)

var akb []byte
var bkb []byte
var amb []byte
var bmb []byte

func TestBytesCmp(t *testing.T) {
	akb = make([]byte, 1024)
	bkb = make([]byte, 1024)
	amb = make([]byte, 1024*1024)
	bmb = make([]byte, 1024*1024)
	if BytesCmp(akb, bkb) == false {
		t.Errorf("bytescmp err a:%v b:%v", akb, bkb)
		return
	}
	if BytesCmp(akb, amb) == true {
		t.Errorf("bytescmp err a:%v b:%v", akb, amb)
		return
	}
}

func Benchmark_BytesCmp_1KB(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BytesCmp(akb, bkb)
	}
}

func Benchmark_BytesCmp_1MB(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BytesCmp(amb, bmb)
	}
}

func Benchmark_BytesCmp_1KB_Ref(b *testing.B) {
	for i := 0; i < b.N; i++ {
		reflect.DeepEqual(akb, bkb)
	}
}

func Benchmark_BytesCmp_1MB_Ref(b *testing.B) {
	for i := 0; i < b.N; i++ {
		reflect.DeepEqual(amb, bmb)
	}
}

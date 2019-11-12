package cache

import (
	"fmt"
	"mmapcache/byteio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var pwd string
var template []byte
var cachesize = 1024 * 1024 * 1
var datasize = 1024 * 8

func TestCreateMMapCache(t *testing.T) {
	pwd, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	template = createMMapTemplate(cachesize)

	cachefile := fmt.Sprintf("%v/0.dat", pwd)
	t.Logf("cachefile:%v", cachefile)

	err := createMMapFile(cachefile, template)
	if nil != err {
		t.Errorf("createMMapFile failed err:%v", err)
		return
	}
	t.Logf("createMMapFile is ok %v", cachefile)

	mmapCache, err := newMMapCache(cachefile, datasize, false)
	if nil != err {
		t.Errorf("newMMapCache failed err:%v", err)
		return
	}
	if len(mmapCache.buf) != cachesize {
		t.Errorf("newMMapCache size:%v err(alloc:%v)", len(mmapCache.buf), cachesize)
		return
	}
	t.Logf("mmapcache.new is ok")

	mmapCache.close(true)
	_, err = os.Stat(cachefile)
	if nil == err || !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("mmapcache.close(ture) failed. err:%v", err)
		return
	}
	t.Logf("mmapcache.close is ok err:%v", err)
}

func TestMMapCacheRecyle(t *testing.T) {
	cachefile := fmt.Sprintf("%v/0.dat", pwd)
	t.Logf("cachefile:%v", cachefile)

	createMMapFile(cachefile, template)
	mmapCache, _ := newMMapCache(cachefile, datasize, false)
	defer mmapCache.close(true)

	// recyle
	mmapCache.recycle(template)
	if mmapCache.getWritePos() != 0 {
		t.Errorf("mmapcache.recycle buf.writepos:%v err", mmapCache.getWritePos())
		return
	}
	if mmapCache.writePos != 0 {
		t.Errorf("mmapcache.recycle writepos:%v err", mmapCache.writePos)
		return
	}
	if mmapCache.readPos != 0 {
		t.Errorf("mmapcache.recycle readpos:%v err", mmapCache.readPos)
		return
	}
	t.Logf("mmapcache.recyle is ok")
}

func TestMMapCacheWrite(t *testing.T) {
	cachefile := fmt.Sprintf("%v/0.dat", pwd)
	t.Logf("cachefile:%v", cachefile)

	createMMapFile(cachefile, template)
	mmapCache, _ := newMMapCache(cachefile, datasize, false)
	defer mmapCache.close(true)

	writeKey := "HelloMMap"
	writeBuf := []byte(writeKey)
	t.Logf("mmapcache.write key:%v buf:%v", writeKey, writeBuf)

	// write.check
	n, err := mmapCache.WriteData(0x1, make([]byte, datasize*2), []byte(writeKey), nil)
	if 0 != n || nil == err {
		t.Errorf("mmapcache.write check err. n:%v err:%v", n, err)
		return
	}
	t.Logf("mmapcache.write check ok")

	// write
	oldDataLen := mmapCache.getFreeContentLen()
	n, _ = mmapCache.WriteData(0xabcd, writeBuf, []byte(writeKey), nil)
	newDataLen := mmapCache.getFreeContentLen()
	if n != len(writeBuf) {
		t.Errorf("mmapcache.write err. n:%v input:%v", n, writeBuf)
		return
	}
	if oldDataLen-mmapCache.dataSize != newDataLen {
		t.Errorf("mmapcache.write err. old.len:%v datasize:%v -> %v != new.len:%v",
			oldDataLen, mmapCache.dataSize, oldDataLen-mmapCache.dataSize, newDataLen)
		return
	}
	t.Logf("mmapcache.write ok. old.len:%v n:%v -> %v != new.len:%v",
		oldDataLen, n, oldDataLen-n-mmapDataHeadLen, newDataLen)

	// mmapdate
	mmapData, _ := mmapCache.mmapdataIdx[writeKey]
	if nil == mmapData {
		t.Errorf("mmapcache.mmapdata idx not found")
		return
	}
	if mmapData.GetSize() != uint32(datasize) {
		t.Errorf("mmapcache.mmapdata size:%v != %v", mmapData.GetSize(), datasize)
		return
	}
	if mmapData.getKeyUsed() != uint32(len([]byte(writeKey))) {
		t.Errorf("mmapcache.mmapdata key.used:%v != %v", mmapData.getKeyUsed(), uint32(len([]byte(writeKey))))
		return
	}
	if mmapData.getDataUsed() != uint32(n) {
		t.Errorf("mmapcache.mmapdata data.used:%v != %v", mmapData.getDataUsed(), n)
		return
	}
	if mmapData.getUsed() != mmapData.getKeyUsed()+mmapData.getDataUsed() {
		t.Errorf("mmapcache.mmapdata used:%v != %v", mmapData.getUsed(), mmapData.getKeyUsed()+mmapData.getDataUsed())
		return
	}
	if mmapData.GetTag() != 0xabcd {
		t.Errorf("mmapcache.mmapdata tag:%v != %v head:%v", mmapData.GetTag(), 0xabcd, mmapData.getHead())
		return
	}
	if false == byteio.BytesCmp(mmapData.GetData(), writeBuf) {
		t.Errorf("mmapcache.mmapdata data:%v != %v", mmapData.GetData()[:len(writeBuf)], writeBuf)
		return
	}
	t.Logf("mmapcache.mmapdata is ok. used:%v tag:%v data:%v",
		mmapData.getUsed(), mmapData.GetTag(), mmapData.GetData())

	// write.overflow check
	for i := 0; ; i++ {
		key := fmt.Sprintf("key-%v", i)
		n, _ := mmapCache.WriteData(0x2, writeBuf, []byte(key), nil)
		if -1 == n {
			break
		}
	}
	if len(mmapCache.mmapdataAry)*mmapCache.dataSize+mmapCacheHeadSize > len(mmapCache.buf) {
		t.Errorf("mmapcache.write err. overflow mmapdata.count:%v totaldatasize:%v headsize:%v => %v : %v",
			len(mmapCache.mmapdataAry), mmapCache.dataSize, mmapCacheHeadSize,
			len(mmapCache.mmapdataAry)*mmapCache.dataSize+mmapCacheHeadSize, len(mmapCache.buf))
	}
	leftSize := len(mmapCache.buf) - (len(mmapCache.mmapdataAry)*mmapCache.dataSize + mmapCacheHeadSize)
	if leftSize >= mmapCache.dataSize {
		t.Errorf("mmapcache.write err. used err. mmapdata.count:%v totaldatasize:%v headsize:%v => %v : %v",
			len(mmapCache.mmapdataAry), mmapCache.dataSize, mmapCacheHeadSize,
			len(mmapCache.mmapdataAry)*mmapCache.dataSize+mmapCacheHeadSize, len(mmapCache.buf))
	}
	t.Logf("mmapcache.write overflow ok. mmapdata.count:%v totaldatasize:%v headsize:%v => %v - %v => %v",
		len(mmapCache.mmapdataAry), mmapCache.dataSize, mmapCacheHeadSize,
		len(mmapCache.buf),
		len(mmapCache.mmapdataAry)*mmapCache.dataSize+mmapCacheHeadSize,
		leftSize)
}

func TestMMapCacheReload(t *testing.T) {
	cachefile := fmt.Sprintf("%v/1.dat", pwd)
	t.Logf("cachefile:%v", cachefile)

	createMMapFile(cachefile, template)
	mmapCache, _ := newMMapCache(cachefile, datasize, false)

	writeCount := 50
	for i := 0; i < writeCount; i++ {
		key := fmt.Sprintf("key-%v", i)
		data := fmt.Sprintf("data-%v", i)
		_, err := mmapCache.WriteData(uint16(i), []byte(data), []byte(key), nil)
		if nil != err {
			t.Errorf("mmapcache.wirte err:%v", err)
			return
		}
	}
	mmapCache.close(false)

	mmapCache, _ = newMMapCache(cachefile, datasize, true)
	defer mmapCache.close(true)
	if len(mmapCache.mmapdataAry) != writeCount {
		t.Errorf("mmapcache.data len:%v != %v", len(mmapCache.mmapdataAry), writeCount)
		return
	}
	t.Logf("mmapcache.data reload ok len:%v = %v", len(mmapCache.mmapdataAry), writeCount)

	for i, mmapdata := range mmapCache.mmapdataAry {
		key := fmt.Sprintf("key-%v", i)
		data := fmt.Sprintf("data-%v", i)

		if string(mmapdata.GetKey()) != key {
			t.Errorf("mmapcache.data reload err idx:%v key:%v != %v", i, string(mmapdata.GetKey()), key)
			return
		}

		if string(mmapdata.GetData()) != data {
			t.Errorf("mmapcache.data reload err idx:%v data:%v != %v", i, string(mmapdata.GetData()), data)
			return
		}
	}
}

var mmapCacheBench *MMapCache
var fileBench *os.File
var fileCounter int
var key []byte
var data []byte

func TestMMapCacheBenchPrepare(t *testing.T) {
	key = make([]byte, 256)
	data = make([]byte, 4*1024)

	cachefile := fmt.Sprintf("%v/b.dat", pwd)
	createMMapFile(cachefile, template)
	mmapCacheBench, _ = newMMapCache(cachefile, datasize, false)

	filePath := fmt.Sprintf("%v/bf.dat", pwd)
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if nil != err {
		t.Errorf("os.OpenFile err:%v", err)
		return
	}
	fileBench = f
}

func Benchmark_MMapcache_Create(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cachefile := fmt.Sprintf("%v/%v.dat", pwd, fileCounter)
		fileCounter++
		createMMapFile(cachefile, template)
	}
}

func Benchmark_MMapcache_Write(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n, _ := mmapCacheBench.WriteData(0x1, data, key, nil)
		if -1 == n {
			mmapCacheBench.recycle(template)
			mmapCacheBench.WriteData(0x1, data, key, nil)
		}
	}
}

func Benchmark_File_Write(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fileBench.Write(key)
		fileBench.Write(data)
	}
}

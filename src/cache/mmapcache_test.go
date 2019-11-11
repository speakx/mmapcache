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
var cachesize = 1024 * 1024 * 10

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

	mmapCache, err := newMMapCache(cachefile, 1024)
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

func TestMMapCache(t *testing.T) {
	cachefile := fmt.Sprintf("%v/0.dat", pwd)
	t.Logf("cachefile:%v", cachefile)

	createMMapFile(cachefile, template)
	mmapCache, _ := newMMapCache(cachefile, 1024)
	defer mmapCache.close(true)

	id := makeMMapCacheID(0x1)
	mmapCache.name(id)
	t.Logf("mmapcache.id:%x", id)
	if mmapCache.GetID() != id {
		t.Errorf("mmapcache.name failed")
		return
	}
	t.Logf("mmapcache.name is ok")

	// merge
	cachefile2 := fmt.Sprintf("%v/1.dat", pwd)
	createMMapFile(cachefile2, template)
	mmapCache2, _ := newMMapCache(cachefile2, 1024)
	defer mmapCache2.close(true)

	id2 := makeMMapCacheID(0x2)
	mmapCache2.name(id2)
	t.Logf("mmapcache2.id:%x", id2)
	mmapCacheMerge := mmapCache.MergeMMapCache(mmapCache2)
	if mmapCacheMerge != mmapCache2 {
		t.Errorf("mmapcache.merge failed. 1:%p 2:%p -> %p", mmapCache, mmapCache2, mmapCacheMerge)
		return
	}
	if mmapCacheMerge.nextMMapCache != mmapCache {
		t.Errorf("mmapcache.merge failed. next:%p  != %p", mmapCacheMerge.nextMMapCache, mmapCache)
		return
	}
	if mmapCacheMerge.GetNextID() != mmapCache.GetID() {
		t.Errorf("mmapcache.merge failed. next.id:%v  != %v", mmapCacheMerge.GetNextID(), mmapCache.GetID())
		return
	}
	t.Logf("mmapcache.merge 1:%p 2:%p -> %p", mmapCache, mmapCache2, mmapCacheMerge)
	t.Logf("mmapcache.merge id:%x nextid:%x", mmapCacheMerge.GetID(), mmapCacheMerge.GetNextID())

	// recyle
	mmapCacheMerge.recycle(template)
	if mmapCacheMerge.getWritePos() != 0 {
		t.Errorf("mmapcache.recycle writepos:%v err", mmapCacheMerge.getWritePos())
		return
	}
	if mmapCacheMerge.readPos != 0 {
		t.Errorf("mmapcache.recycle readpos:%v err", mmapCacheMerge.readPos)
		return
	}
	if mmapCacheMerge.readPos != 0 {
		t.Errorf("mmapcache.recycle readpos:%v err", mmapCacheMerge.readPos)
		return
	}
	if mmapCacheMerge.nextMMapCache != nil {
		t.Errorf("mmapcache.recycle next:%v err", mmapCacheMerge.nextMMapCache)
		return
	}
	if mmapCacheMerge.GetNextID() != 0 {
		t.Errorf("mmapcache.recycle nextid:%v err", mmapCacheMerge.GetNextID())
		return
	}
	t.Logf("mmapcache.recyle is ok")
}

func TestMMapCacheWrite(t *testing.T) {
	cachefile := fmt.Sprintf("%v/0.dat", pwd)
	t.Logf("cachefile:%v", cachefile)

	createMMapFile(cachefile, template)
	mmapCache, _ := newMMapCache(cachefile, 1024)
	defer mmapCache.close(true)

	writeKey := "HelloMMap"
	writeBuf := []byte(writeKey)
	t.Logf("mmapcache.write key:%v buf:%v", writeKey, writeBuf)

	// write.check
	n, err := mmapCache.Write(0x1, make([]byte, 1024*2), []byte(writeKey), nil)
	if 0 != n || nil == err {
		t.Errorf("mmapcache.write check err. n:%v err:%v", n, err)
		return
	}
	t.Logf("mmapcache.write check ok")

	// write
	oldDataLen := mmapCache.GetFreeDataLen()
	n, _ = mmapCache.Write(0xabcd, writeBuf, []byte(writeKey), nil)
	newDataLen := mmapCache.GetFreeDataLen()
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
	if mmapData.getUsed() != uint32(n) {
		t.Errorf("mmapcache.mmapdata used:%v != %v", mmapData.getUsed(), n)
		return
	}
	if mmapData.getTag() != 0xabcd {
		t.Errorf("mmapcache.mmapdata tag:%v != %v head:%v", mmapData.getTag(), 0xabcd, mmapData.getHead())
		return
	}
	if false == byteio.BytesCmp(mmapData.getData(), writeBuf) {
		t.Errorf("mmapcache.mmapdata data:%v != %v", mmapData.getData()[:len(writeBuf)], writeBuf)
		return
	}
	t.Logf("mmapcache.mmapdata is ok. used:%v tag:%v data:%v",
		mmapData.getUsed(), mmapData.getTag(), mmapData.getData())

	// write.overflow check
	for i := 0; ; i++ {
		key := fmt.Sprintf("key-%v", i)
		n, _ := mmapCache.Write(0x2, writeBuf, []byte(key), nil)
		if -1 == n {
			break
		}
	}
	if len(mmapCache.mmapdataAry)*mmapCache.dataSize+mmapDataHeadLen > len(mmapCache.buf) {
		t.Errorf("mmapcache.write err. overflow mmapdata.count:%v totaldatasize:%v headsize:%v => %v : %v",
			len(mmapCache.mmapdataAry), mmapCache.dataSize, mmapDataHeadLen,
			len(mmapCache.mmapdataAry)*mmapCache.dataSize+mmapDataHeadLen, len(mmapCache.buf))
	}
	leftSize := len(mmapCache.buf) - (len(mmapCache.mmapdataAry)*mmapCache.dataSize + mmapDataHeadLen)
	if leftSize >= mmapCache.dataSize {
		t.Errorf("mmapcache.write err. used err. mmapdata.count:%v totaldatasize:%v headsize:%v => %v : %v",
			len(mmapCache.mmapdataAry), mmapCache.dataSize, mmapDataHeadLen,
			len(mmapCache.mmapdataAry)*mmapCache.dataSize+mmapDataHeadLen, len(mmapCache.buf))
	}
	t.Logf("mmapcache.write overflow ok. mmapdata.count:%v totaldatasize:%v headsize:%v => %v : %v",
		len(mmapCache.mmapdataAry), mmapCache.dataSize, mmapDataHeadLen,
		len(mmapCache.mmapdataAry)*mmapCache.dataSize+mmapDataHeadLen, len(mmapCache.buf))
}

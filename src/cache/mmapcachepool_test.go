package cache

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
)

var poolpwd string
var poolcachesize = 1024 * 1024 * 1
var pooldatasize = 1024 * 8
var poolcnt = 100

func TestPoolMMapCache(t *testing.T) {
	poolpwd, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	ioutil.WriteFile(
		path.Join(poolpwd, "x.cachedat.err"),
		make([]byte, poolcachesize), 0666)

	InitMMapCachePool(
		poolpwd, poolcachesize, pooldatasize, poolcnt,
		func(err error) {
			fmt.Printf("poolmmapcache err:%v\n", err)
		},
		func(mmapCaches []*MMapCache) {
		})

	for index := 0; index < 10; index++ {
		mmapCache := DefPoolMMapCache.Alloc()
		t.Logf("alloc mmapcache %p file:%v", mmapCache, mmapCache.path)
		key := []byte(fmt.Sprintf("key-%v", index))
		data := []byte(fmt.Sprintf("data-%v", index))

		mmapCache.WriteData(uint16(index), data, key, nil)
		mmapCache.close(false)
	}
	DefPoolMMapCache.close()
	t.Logf("mmapcache.pool closed")

	InitMMapCachePool(
		poolpwd, poolcachesize, pooldatasize, poolcnt,
		func(err error) {
			fmt.Printf("poolmmapcache err:%v\n", err)
		},
		func(mmapCaches []*MMapCache) {
			for _, mmapCache := range mmapCaches {
				t.Logf("reload mmapcache %p file:%v", mmapCache, mmapCache.path)
				t.Logf(" -- data.len:%v key.0:%v val.0:%v",
					len(mmapCache.GetMMapDatas()),
					string(mmapCache.GetMMapDatas()[0].GetKey()),
					string(mmapCache.GetMMapDatas()[0].GetData()))
			}
		})
}

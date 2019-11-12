package main

import (
	"encoding/json"
	"fmt"
	"mmapcache/cache"
	"os"
	"path/filepath"
	"time"
)

type keyVal struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

func main() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	fmt.Println(dir)
	cache.InitMMapCachePool(
		dir, 1024*1024*1, 128, 5,
		func(err error) {

		},
		func(mmapCaches []*cache.MMapCache) {
			fmt.Printf("used mmapcaches.len:%v\n", len(mmapCaches))
			for _, mmapCache := range mmapCaches {
				fmt.Printf("reload mmap file %v\n", mmapCache.Path())
				for _, mmapData := range mmapCache.GetMMapDatas() {
					var vk *keyVal
					json.Unmarshal(mmapData.GetData(), &vk)
					fmt.Println(vk)
					mmapData.ReloadVal(vk)
				}
			}

			for _, mmapCache := range mmapCaches {
				mmapCache.Release()
			}
		})

	mmapCache := cache.DefPoolMMapCache.Alloc()
	fmt.Printf("alloc:%v\n", mmapCache.Path())
	for i := 0; i < 10; i++ {
		vk := &keyVal{
			Key: fmt.Sprintf("Key-%v", i),
			Val: fmt.Sprintf("Val-%v", i),
		}
		data, _ := json.Marshal(vk)
		mmapCache.WriteData(0x0, data, []byte(vk.Key), vk)
	}

	chunkBuf := mmapCache.GetWrittenData()
	reloadMMapCache := cache.ReloadMMapCache(chunkBuf)
	for _, mmapData := range reloadMMapCache.GetMMapDatas() {
		var vk *keyVal
		json.Unmarshal(mmapData.GetData(), &vk)
		fmt.Printf("reload memory %v\n", vk)
		mmapData.ReloadVal(vk)
	}

	wait := 10
	for i := 0; i < wait; i++ {
		<-time.After(time.Second)
		alloc, collect, release, size := cache.DefPoolMMapCache.DumpRuntime()
		fmt.Printf("time.after %v cache.pool alloc:%v collect:%v release:%v size:%v\n",
			wait-i, alloc, collect, release, size)
	}
}

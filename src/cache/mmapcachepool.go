package cache

import (
	"container/list"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

// DefPoolMMapCache Init初始化完成后，得到的内存池对象
var DefPoolMMapCache *PoolMMapCache

const mmapInitByte byte = 0

// PoolMMapCache 通过mmap方式对内存对象持久化缓存
type PoolMMapCache struct {
	dir            string
	template       []byte
	dataSize       int
	pool           *list.List
	recycleDur     time.Duration
	allocator      chan *MMapCache
	collector      chan *MMapCache
	errorfuc       func(error)
	loadFlag       bool
	closedFlag     bool
	allocCounter   uint64
	collectCounter uint64
	releaseCounter uint64
	poolSize       int
	wait           sync.WaitGroup
}

// InitMMapCachePool 初始化mmap的cache池
// mmapsize 缓存文件大小
// datasize 缓存数据块大小
// prealloc 初始化缓存池时，会预先构建的缓存文件数量
// errorfunc 当出现异常，会出发此函数异步抛出error
// reloadfunc 当本地有之前的缓存数据时，通过此函数处理已经缓存到本地的数据
//            这批数据会初始化为MMapCache对象，但不会被添加到缓存池中（因为其中已经有数据，处于正在使用状态）
// func(mmapCaches []*MMapCache) {
// 	for _, mmapCache := range mmapCaches {
// 		for _,mmapData := range mmapCache.GetMMapDatas() {
// 			val := 反序列化文件缓存数据(mmapData.GetData())
// 			mmapData.ReloadVal(val)
// 		}
// 	}
// }
func InitMMapCachePool(
	dir string,
	mmapsize int, datasize int, prealloc int,
	errorfunc func(error),
	reloadfunc func([]*MMapCache)) error {
	os.MkdirAll(dir, os.ModePerm)
	DefPoolMMapCache = &PoolMMapCache{
		dir:        dir,
		template:   createMMapTemplate(mmapsize),
		dataSize:   datasize,
		pool:       list.New(),
		recycleDur: time.Second,
		allocator:  make(chan *MMapCache),
		collector:  make(chan *MMapCache),
		errorfuc:   errorfunc,
		loadFlag:   false,
	}

	reload := DefPoolMMapCache.reloadCache()
	DefPoolMMapCache.wait.Add(1)
	DefPoolMMapCache.mmapAllocLoop(prealloc)
	for {
		if DefPoolMMapCache.loadFlag == false {
			<-time.After(time.Millisecond * 50)
		} else {
			break
		}
	}

	reloadfunc(reload)
	return nil
}

// Alloc 分配一个mmapcache
func (m *PoolMMapCache) Alloc() *MMapCache {
	return <-m.allocator
}

// Collect 回收一个mmapcache到缓存池
func (m *PoolMMapCache) Collect(mmcache *MMapCache) {
	m.collector <- mmcache
}

// DumpRuntime 获取缓存池当前的数据指标
// AllocCounter, CollectCounter, ReleaseCounter, PoolSize
func (m *PoolMMapCache) DumpRuntime() (uint64, uint64, uint64, int) {
	return m.allocCounter, m.collectCounter, m.releaseCounter, m.poolSize
}

func createMMapTemplate(size int) []byte {
	template := make([]byte, size)
	for index := 0; index < size; index++ {
		template[index] = mmapInitByte
	}
	return template
}

func createMMapFile(file string, template []byte) error {
	return ioutil.WriteFile(file, template, 0666)
}

func (m *PoolMMapCache) makeCacheFileName() string {
	fileName := fmt.Sprintf("%v_%v", uint32(time.Now().Unix()), m.allocCounter)
	m.allocCounter++
	return path.Join(m.dir, fmt.Sprintf("%v.cachedat", fileName))
}

func (m *PoolMMapCache) close() {
	m.closedFlag = true
	m.wait.Wait()
}

func (m *PoolMMapCache) reloadCache() []*MMapCache {
	fis, err := ioutil.ReadDir(m.dir)
	if err != nil {
		return nil
	}

	reloadMMapCaches := make([]*MMapCache, 0, 10)
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}

		ok := strings.HasSuffix(fi.Name(), ".cachedat")
		if ok {
			filePath := path.Join(m.dir, fi.Name())

			mmapCache, err := newMMapCache(filePath, m.dataSize, true)
			// 数据没发加载，移动为.err文件，待分析
			if nil != err {
				os.Rename(filePath, fmt.Sprintf("%v.err", filePath))
				continue
			}

			// 有数据，加入到reload队列抛给业务层自行处理
			if mmapCache.getWritePos() > 0 {
				reloadMMapCaches = append(reloadMMapCaches, mmapCache)
				continue
			}

			// 没有数据，判断一下文件大小是否一样，不一样就删了
			if int(fi.Size()) != len(m.template) {
				mmapCache.close(true)
				continue
			}

			// 啥问题都没，加到缓存池里
			m.pool.PushBack(mmapCache)
		}
	}
	return reloadMMapCaches
}

func (m *PoolMMapCache) preAllocMMapCache() *MMapCache {
	filePath := m.makeCacheFileName()

	createFlag := false
	fi, _ := os.Stat(filePath)
	if nil != fi {
		if int(fi.Size()) < len(m.template) {
			os.Remove(filePath)
			createFlag = true
		}
	} else {
		createFlag = true
	}

	if createFlag {
		err := createMMapFile(filePath, m.template)
		if nil != err {
			m.errorfuc(err)
			os.Exit(0)
		}
	}

	mmapCache, err := newMMapCache(filePath, m.dataSize, false)
	if nil != err {
		m.errorfuc(err)
		os.Exit(0)
	}
	return mmapCache
}

func (m *PoolMMapCache) mmapAllocLoop(cnt int) {
	go func() {
		for {
			if m.pool.Len() < cnt {
				m.pool.PushBack(m.preAllocMMapCache())
			} else {
				break
			}
		}
		m.loadFlag = true

		for {
			if m.closedFlag {
				for {
					if m.pool.Len() == 0 {
						break
					}
					e := m.pool.Front()
					e.Value.(*MMapCache).close(false)
					m.pool.Remove(e)
				}
				break
			}

			if m.pool.Len() < cnt/2 {
				m.pool.PushBack(m.preAllocMMapCache())
			}

			m.poolSize = m.pool.Len()
			e := m.pool.Front()

			select {
			case b := <-m.collector:
				if m.pool.Len() < cnt*2 {
					b.recycle(m.template)
					m.pool.PushBack(b)
					m.collectCounter++
				} else {
					b.close(true)
				}
			case m.allocator <- e.Value.(*MMapCache):
				m.pool.Remove(e)
			case <-time.After(m.recycleDur):
				if m.pool.Len() > cnt {
					for i := 0; i < 10 && m.pool.Len() > cnt; i++ {
						e := m.pool.Back()
						e.Value.(*MMapCache).close(true)
						m.pool.Remove(e)
						m.releaseCounter++
					}
				} else {
					break
				}
			}
		}
		m.wait.Done()
	}()
}

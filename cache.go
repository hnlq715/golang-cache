package cache

import (
	"bytes"
	"cache/file"
	"crypto/md5"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/golang/glog"

	"gopkg.in/redis.v4"
)

type Option struct {
	Expire    time.Duration
	Valid     time.Duration
	UseStale  int32
	LockTimes int32

	File *file.Option

	RedisType    string
	RedisCluster *redis.ClusterOptions
	RedisRing    *redis.RingOptions
	RedisExpire  time.Duration
}

type cacheInfo struct {
	mu        sync.RWMutex
	lockTimes int32
	use       int64
	timestamp time.Time
}

type Cache struct {
	option *Option

	mu   sync.RWMutex
	info map[string]*cacheInfo

	nbytes     AtomicInt
	fhit, fget AtomicInt
	rhit, rget AtomicInt
	nevicted   AtomicInt

	file    *file.File
	cluster *redis.ClusterClient
	ring    *redis.Ring
}

const MISS = "MISS"
const BYPASS = "BYPASS"
const EXPIRED = "EXPIRED"
const STALE = "STALE"
const UPDATING = "UPDATING"
const REVALIDATED = "REVALIDATED"

const RedisCluster = "cluster"
const RedisRing = "ring"

const defaultFileDirName = "data"

func New(option *Option) *Cache {
	c := new(Cache)

	c.info = make(map[string]*cacheInfo)
	if option != nil {
		c.option = option

		if option.RedisCluster != nil {
			c.cluster = redis.NewClusterClient(option.RedisCluster)
		} else if option.RedisRing != nil {
			c.ring = redis.NewRing(option.RedisRing)
		}
	}

	c.file = file.New(&file.Option{
		DirName: defaultFileDirName,
	})

	go c.load()

	return c
}

func (c *Cache) Get(key string) ([]byte, error) {
	ok := c.has(key)
	if !ok {
		data, err := c.getFromRedisAndSetFile(key)
		if err != nil {
			glog.Errorf("c.getFromRedisAndSetFile failed, err=%s", err)
			return nil, err
		}

		return data, err
	}

	data, err := c.getFile(key)

	if c.Expired(key) {
		data, err := c.getFromRedis(key)
		if err != nil {
			if err != redis.Nil {
				return data, err
			}
			glog.Errorf("c.getFromRedis failed, err=%s", err)
			return nil, err
		}

		return data, err
	}

	return data, err
}

func (c *Cache) Set(key string, data []byte) error {
	err := c.setFile(key, data)
	if err != nil {
		glog.Errorf("c.setFile failed, err=%s", err)
		return err
	}

	err = c.setRedis(key, data)
	if err != nil {
		glog.Errorf("c.setRedis failed, err=%s", err)
		return err
	}

	return nil
}

func (c *Cache) Delete(key string) error {
	filename := filepath.Join(defaultFileDirName, generateFileName(key))

	if err := os.Remove(filename); err != nil {
		return err
	}

	return nil
}

func (c *Cache) getFile(key string) ([]byte, error) {
	data, err := c.file.Get(key)
	if err == nil {
		c.fhit.Add(1)
	}
	c.fget.Add(1)

	return data, err
}

func (c *Cache) setFile(key string, data []byte) error {
	err := c.file.Set(key, bytes.NewReader(data))
	if err != nil {
		glog.Errorf("c.file.Set failed, err=%s", err)
		return err
	}

	// c.mu.Lock()
	// info, ok := c.info[key]
	// if ok {
	// 	info.timestamp = time.Now()
	// } else {
	// 	c.info[key] = &cacheInfo{timestamp: time.Now()}
	// }
	// c.mu.Unlock()

	c.nbytes.Add(int64(len(data)))

	return nil
}

func (c *Cache) getFromRedis(key string) ([]byte, error) {
	var g Group
	v, err := g.Do(key, func() (interface{}, error) {
		data, err := c.getRedis(key)
		if err != nil {
			glog.Errorf("c.getRedis failed, err=%s", err)
			return nil, err
		}
		return data, err
	})

	return v.([]byte), err
}

func (c *Cache) getFromRedisAndSetFile(key string) ([]byte, error) {
	var g Group
	v, err := g.Do(key, func() (interface{}, error) {
		data, err := c.getRedis(key)
		if err != nil {
			glog.Errorf("c.getRedis failed, err=%s, key=%s", err, key)
			return data, err
		}
		c.setFile(key, data)
		return data, err
	})

	return v.([]byte), err
}

func (c *Cache) getRedis(key string) ([]byte, error) {
	var data []byte
	var err error
	if c.cluster != nil {
		data, err = c.cluster.Get(key).Bytes()
	} else if c.ring != nil {
		data, err = c.ring.Get(key).Bytes()
	}

	c.rget.Add(1)
	if err == nil {
		c.rhit.Add(1)
	}

	return data, err
}

func (c *Cache) setRedis(key string, data []byte) error {
	if c.cluster != nil {
		return c.cluster.Set(key, data, c.option.RedisExpire).Err()
	} else if c.ring != nil {
		return c.ring.Set(key, data, c.option.RedisExpire).Err()
	}

	return errors.New("no redis client found")
}

func (c *Cache) load() {
	dir, _ := filepath.Abs(defaultFileDirName)

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		basename := filepath.Base(path)

		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		if len(basename) != md5.Size {
			os.Remove(path)
			return nil
		}

		c.updateCacheInfo(basename)

		return nil
	})
}

func (c *Cache) updateCacheInfo(key string) {
	c.mu.Lock()
	info, ok := c.info[key]
	if ok {
		info.timestamp = time.Now()
		info.use++
	} else {
		info = &cacheInfo{timestamp: time.Now()}
	}
	c.mu.Unlock()
}

func (c *Cache) has(key string) bool {
	c.mu.RLock()
	_, ok := c.info[key]
	c.mu.RUnlock()
	return ok
}

func (c *Cache) Expired(key string) bool {
	c.mu.RLock()
	flag := time.Now().Sub(c.info[key].timestamp) > c.option.Expire
	c.mu.RUnlock()
	return flag
}

func init() {
	flag.Parse()
}

package cache

import (
	"errors"
	"flag"
	"time"

	"github.com/golang/glog"
	"github.com/hnlq715/golang-lru"

	"gopkg.in/redis.v4"
)

type Option struct {
	LRU   *LRUOption
	Redis *RedisOption
}

type LRUOption struct {
	MaxSize int
	Expire  time.Duration
}

type RedisOption struct {
	Cluster *redis.ClusterOptions
	Ring    *redis.RingOptions
	Expire  time.Duration
}

type Cache struct {
	option *Option

	g Group

	stats CacheStats

	arc     *lru.ARCCache
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

func New(option *Option) *Cache {
	c := new(Cache)

	c.option = option

	if option != nil {
		if option.LRU != nil {
			c.arc, _ = lru.NewARC(option.LRU.MaxSize)
		}

		if option.Redis != nil {
			if option.Redis.Cluster != nil {
				c.cluster = redis.NewClusterClient(option.Redis.Cluster)
			} else if option.Redis.Ring != nil {
				c.ring = redis.NewRing(option.Redis.Ring)
			}
		}
	}

	return c
}

func (c *Cache) Get(key string) ([]byte, error) {
	ok := c.arc.Contains(key)
	if !ok {
		data, err := c.getFromRedis(key)
		if err != nil {
			glog.Errorf("c.getFromRedisAndSetFile failed, err=%s", err)
			return nil, err
		}

		c.arc.AddWithExpire(key, data, c.option.LRU.Expire)
		return data, err
	}

	c.stats.LGets.Add(1)
	data, ok := c.arc.Get(key)
	if !ok {
		return nil, errors.New("c.arc.Get failed")
	}
	c.stats.LHits.Add(1)

	return data.([]byte), nil
}

func (c *Cache) Set(key string, data []byte) error {

	err := c.setRedis(key, data, c.option.Redis.Expire)
	if err != nil {
		glog.Errorf("c.setRedis failed, err=%s", err)
		return err
	}

	c.arc.AddWithExpire(key, data, c.option.LRU.Expire)

	return nil
}

func (c *Cache) getFromRedis(key string) ([]byte, error) {
	v, err := c.g.Do(key, func() (interface{}, error) {
		data, err := c.getRedis(key)
		if err != nil {
			glog.Errorf("c.getRedis failed, err=%s", err)
			return nil, err
		}
		return data, err
	})
	if err != nil {
		return nil, err
	}

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

	c.stats.RGets.Add(1)
	if err == nil {
		c.stats.RHits.Add(1)
	}

	return data, err
}

func (c *Cache) setRedis(key string, data []byte, expire time.Duration) error {
	if c.cluster != nil {
		return c.cluster.Set(key, data, expire).Err()
	} else if c.ring != nil {
		return c.ring.Set(key, data, expire).Err()
	}

	return errors.New("no redis client found")
}

func (c *Cache) Stats() *CacheStats {
	return &c.stats
}

func init() {
	flag.Parse()
}

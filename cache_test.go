package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/redis.v4"
)

func TestSetAndGet(t *testing.T) {
	c := New(&Option{
		LRU: &LRUOption{
			Expire:  300 * time.Second,
			MaxSize: 100,
		},
		Redis: &RedisOption{
			Expire: 600 * time.Second,
			Ring: &redis.RingOptions{
				Addrs: map[string]string{
					"redis01": "192.168.1.77:6379",
				},
			},
		},
	})

	err := c.Set("test", []byte("we are here"))
	assert.Nil(t, err)
	val, err := c.Get("test")
	assert.Nil(t, err)
	assert.Equal(t, []byte("we are here"), val)
}

func BenchmarkSet(b *testing.B) {
	c := New(&Option{
		LRU: &LRUOption{
			Expire:  300 * time.Second,
			MaxSize: 100,
		},
		Redis: &RedisOption{
			Expire: 600 * time.Second,
			Ring: &redis.RingOptions{
				Addrs: map[string]string{
					"redis01": "192.168.1.77:6379",
				},
			},
		},
	})

	time.Sleep(1 * time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := c.Set("test", []byte("we are here"))
		assert.Nil(b, err)
	}
}

func BenchmarkGet(b *testing.B) {
	c := New(&Option{
		LRU: &LRUOption{
			Expire:  300 * time.Second,
			MaxSize: 100,
		},
		Redis: &RedisOption{
			Expire: 600 * time.Second,
			Ring: &redis.RingOptions{
				Addrs: map[string]string{
					"redis01": "192.168.1.77:6379",
				},
			},
		},
	})

	time.Sleep(1 * time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.Get("test")
		assert.Nil(b, err)
	}
}

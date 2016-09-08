package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/redis.v4"
)

func TestSetAndGet(t *testing.T) {
	c := New(&Option{
		Expire:      60 * time.Second,
		LockTimes:   3,
		RedisExpire: 600 * time.Second,
		RedisRing: &redis.RingOptions{
			Addrs: map[string]string{
				"redis01": "192.168.33.10:6379",
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
		Expire:      60 * time.Second,
		LockTimes:   3,
		RedisExpire: 600 * time.Second,
		RedisRing: &redis.RingOptions{
			Addrs: map[string]string{
				"redis01": "192.168.1.77:6379",
			},
		},
	})

	c.load()
	time.Sleep(1 * time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := c.Set("test", []byte("we are here"))
		assert.Nil(b, err)
	}
}

func BenchmarkGet(b *testing.B) {
	c := New(&Option{
		Expire:      60 * time.Second,
		LockTimes:   3,
		RedisExpire: 600 * time.Second,
		RedisRing: &redis.RingOptions{
			Addrs: map[string]string{
				"redis01": "192.168.1.77:6379",
			},
		},
	})

	c.load()
	time.Sleep(1 * time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.Get("test")
		assert.Nil(b, err)
	}
}

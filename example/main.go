package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/hnlq715/cache"

	"gopkg.in/redis.v4"
)

func main() {
	c := cache.New(&cache.Option{
		LRU: &cache.LRUOption{
			Expire:  600 * time.Second,
			MaxSize: 1,
		},
		Redis: &cache.RedisOption{
			Expire: 600 * time.Second,
			Ring: &redis.RingOptions{
				Addrs: map[string]string{
					"redis01": "192.168.1.77:6379",
				},
			},
		},
	})

	http.HandleFunc("/set", func(rw http.ResponseWriter, r *http.Request) {
		err := c.Set("test", []byte("hello worldsdfsdfsdfs"))
		if err != nil {
			log.Fatalf("c.Set failed, err=%s", err)
		}
		rw.Write([]byte("hello worldsdfsdfsdfs"))
	})

	http.HandleFunc("/get", func(rw http.ResponseWriter, r *http.Request) {
		data, err := c.Get("test")
		if err != nil && err != redis.Nil {
			log.Fatalf("c.Get failed, err=%s", err)
		}
		rw.Write(data)
	})

	http.ListenAndServe(":8080", nil)
}

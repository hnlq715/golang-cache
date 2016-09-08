package main

import (
	"cache"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"gopkg.in/redis.v4"
)

func main() {
	c := cache.New(&cache.Option{
		Expire:      60 * time.Second,
		LockTimes:   3,
		RedisExpire: 600 * time.Second,
		RedisRing: &redis.RingOptions{
			Addrs: map[string]string{
				"redis01": "192.168.33.10:6379",
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
		if err != nil {
			log.Fatalf("c.Get failed, err=%s", err)
		}
		rw.Write(data)
	})

	http.ListenAndServe(":8080", nil)
}

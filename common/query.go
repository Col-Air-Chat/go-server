package common

import (
	"time"

	"github.com/fanjindong/go-cache"
)

var (
	c cache.ICache
)

func InitCache() {
	c = cache.NewMemCache()
}

func GetCache() cache.ICache {
	if c == nil {
		InitCache()
	}
	return c
}

func GetDataFromCache(resultChan chan interface{}, found chan bool, key string) {
	defer close(resultChan)
	defer close(found)
	val, ok := c.Get(key)
	found <- ok
	if ok {
		resultChan <- val
	}
}

func CacheExpire(expire time.Duration) cache.SetIOption {
	return cache.WithEx(expire)
}

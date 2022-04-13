package bigcache

import (
	"errors"
	"time"
)

var ErrNotExist = errors.New("cache key not exist")

type Cacher interface {
	Set(key string, value []byte, ttl time.Duration) error
	Get(key string) ([]byte, error)
	TTL(key string) (time.Duration, error)
	Delete(key string)
}

type CacheItem struct {
	Expires time.Time
	Value   []byte
}

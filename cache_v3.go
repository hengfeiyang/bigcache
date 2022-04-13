package bigcache

import (
	"sync"
	"time"
)

// cacheV3 use sync.map
type cacheV3 struct {
	data sync.Map
}

func NewCacheV3(maxEntrySize int) Cacher {
	return &cacheV3{}
}

func (t *cacheV3) Set(key string, value []byte, ttl time.Duration) error {
	expires := time.Time{}
	if ttl > 0 {
		expires = time.Now().Add(ttl)
	}
	t.data.Store(key, &CacheItem{Expires: expires, Value: value})
	return nil
}

func (t *cacheV3) Get(key string) ([]byte, error) {
	v, ok := t.data.Load(key)
	if !ok {
		return nil, ErrNotExist
	}
	item := v.(*CacheItem)
	if !item.Expires.IsZero() && item.Expires.Before(time.Now()) {
		t.Delete(key)
		return nil, ErrNotExist
	}
	return item.Value, nil
}

func (t *cacheV3) TTL(key string) (time.Duration, error) {
	v, ok := t.data.Load(key)
	if !ok {
		return 0, ErrNotExist
	}
	item := v.(*CacheItem)
	if item.Expires.IsZero() {
		return -1, nil
	}
	if item.Expires.Before(time.Now()) {
		t.Delete(key)
		return 0, ErrNotExist
	}
	return time.Until(item.Expires), nil
}

func (t *cacheV3) Delete(key string) {
	t.data.Delete(key)
}

func (t *cacheV3) Len() int {
	i := 0
	t.data.Range(func(key, value interface{}) bool {
		i++
		return true
	})
	return i
}

package bigcache

import (
	"sync"
	"time"
)

// cacheV2 use map and mutex support concurrent
type cacheV2 struct {
	data map[string]CacheItem
	lock sync.RWMutex
}

func NewCacheV2(maxEntrySize int) Cacher {
	return &cacheV2{
		data: make(map[string]CacheItem, maxEntrySize),
	}
}

func (t *cacheV2) Set(key string, value []byte, ttl time.Duration) error {
	expires := time.Time{}
	if ttl > 0 {
		expires = time.Now().Add(ttl)
	}
	t.lock.Lock()
	t.data[key] = CacheItem{Expires: expires, Value: value}
	t.lock.Unlock()
	return nil
}

func (t *cacheV2) Get(key string) ([]byte, error) {
	t.lock.RLock()
	v, ok := t.data[key]
	t.lock.RUnlock()
	if !ok {
		return nil, ErrNotExist
	}
	if !v.Expires.IsZero() && v.Expires.Before(time.Now()) {
		t.Delete(key)
		return nil, ErrNotExist
	}
	return v.Value, nil
}

func (t *cacheV2) TTL(key string) (time.Duration, error) {
	t.lock.RLock()
	v, ok := t.data[key]
	t.lock.RUnlock()
	if !ok {
		return 0, ErrNotExist
	}
	if v.Expires.IsZero() {
		return -1, nil
	}
	if v.Expires.Before(time.Now()) {
		t.Delete(key)
		return 0, ErrNotExist
	}
	return time.Until(v.Expires), nil
}

func (t *cacheV2) Delete(key string) {
	t.lock.Lock()
	delete(t.data, key)
	t.lock.Unlock()
}

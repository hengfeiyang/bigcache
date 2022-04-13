package bigcache

import "time"

// cacheV1 simple a map, don't support concurrent
type cacheV1 struct {
	data map[string]CacheItem
}

func NewCacheV1(maxEntrySize int) Cacher {
	return &cacheV1{
		data: make(map[string]CacheItem, maxEntrySize),
	}
}

func (t *cacheV1) Set(key string, value []byte, ttl time.Duration) error {
	expires := time.Time{}
	if ttl > 0 {
		expires = time.Now().Add(ttl)
	}
	t.data[key] = CacheItem{Expires: expires, Value: value}
	return nil
}

func (t *cacheV1) Get(key string) ([]byte, error) {
	v, ok := t.data[key]
	if !ok {
		return nil, ErrNotExist
	}
	if !v.Expires.IsZero() && v.Expires.Before(time.Now()) {
		t.Delete(key)
		return nil, ErrNotExist
	}
	return v.Value, nil
}

func (t *cacheV1) TTL(key string) (time.Duration, error) {
	v, ok := t.data[key]
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

func (t *cacheV1) Delete(key string) {
	delete(t.data, key)
}

func (t *cacheV1) Len() int {
	return len(t.data)
}

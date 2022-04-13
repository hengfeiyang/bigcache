package bigcache

import (
	"hash/fnv"
	"math"
	"sync"
	"time"
)

// cacheV4 use concurrent-map
type cacheV4 struct {
	shards  []cacheV4Shard
	bitMask uint64
}

type cacheV4Shard struct {
	data map[string]CacheItem
	lock sync.RWMutex
}

func NewCacheV4(maxEntrySize int, shards int) Cacher {
	if shards == 0 {
		shards = 128
	}
	if shards&(shards-1) != 0 {
		shards = int(math.Pow(2, math.Ceil(math.Log2(float64(shards)))))
	}
	t := &cacheV4{
		shards:  make([]cacheV4Shard, shards+1),
		bitMask: uint64(shards - 1),
	}
	for i := 0; i < shards; i++ {
		t.shards[i].data = make(map[string]CacheItem)
	}
	return t
}

func (t *cacheV4) Set(key string, value []byte, ttl time.Duration) error {
	expires := time.Time{}
	if ttl > 0 {
		expires = time.Now().Add(ttl)
	}
	shardKey := t.getShardKey([]byte(key))
	t.shards[shardKey].lock.Lock()
	t.shards[shardKey].data[key] = CacheItem{Expires: expires, Value: value}
	t.shards[shardKey].lock.Unlock()
	return nil
}

func (t *cacheV4) Get(key string) ([]byte, error) {
	shardKey := t.getShardKey([]byte(key))
	t.shards[shardKey].lock.RLock()
	v, ok := t.shards[shardKey].data[key]
	t.shards[shardKey].lock.RUnlock()
	if !ok {
		return nil, ErrNotExist
	}
	if !v.Expires.IsZero() && v.Expires.Before(time.Now()) {
		t.Delete(key)
		return nil, ErrNotExist
	}
	return v.Value, nil
}

func (t *cacheV4) TTL(key string) (time.Duration, error) {
	shardKey := t.getShardKey([]byte(key))
	t.shards[shardKey].lock.RLock()
	v, ok := t.shards[shardKey].data[key]
	t.shards[shardKey].lock.RUnlock()
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

func (t *cacheV4) Delete(key string) {
	shardKey := t.getShardKey([]byte(key))
	t.shards[shardKey].lock.Lock()
	delete(t.shards[shardKey].data, key)
	t.shards[shardKey].lock.Unlock()
}

var h = fnv.New64a()

func (t *cacheV4) getShardKey(key []byte) int {
	h.Reset()
	h.Write(key)
	i := h.Sum64()
	return int(i & t.bitMask)
}
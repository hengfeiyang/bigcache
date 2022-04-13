package bigcache

import (
	"hash/fnv"
	"math"
	"time"
)

// cacheV4 use concurrent-map
type cacheV4 struct {
	shards  []Cacher
	bitMask uint64
}

func NewCacheV4(maxEntrySize int, shards int) Cacher {
	if shards == 0 {
		shards = 128
	}
	if shards&(shards-1) != 0 {
		shards = int(math.Pow(2, math.Ceil(math.Log2(float64(shards)))))
	}
	t := &cacheV4{
		shards:  make([]Cacher, shards+1),
		bitMask: uint64(shards - 1),
	}
	for i := 0; i < shards; i++ {
		t.shards[i] = NewCacheV2(maxEntrySize / shards)
	}
	return t
}

func (t *cacheV4) Set(key string, value []byte, ttl time.Duration) error {
	shardKey := t.getShardKey([]byte(key))
	return t.shards[shardKey].Set(key, value, ttl)
}

func (t *cacheV4) Get(key string) ([]byte, error) {
	shardKey := t.getShardKey([]byte(key))
	return t.shards[shardKey].Get(key)
}

func (t *cacheV4) TTL(key string) (time.Duration, error) {
	shardKey := t.getShardKey([]byte(key))
	return t.shards[shardKey].TTL(key)
}

func (t *cacheV4) Delete(key string) {
	shardKey := t.getShardKey([]byte(key))
	t.shards[shardKey].Delete(key)
}

var h = fnv.New64a()

func (t *cacheV4) getShardKey(key []byte) int {
	h.Reset()
	h.Write(key)
	i := h.Sum64()
	return int(i & t.bitMask)
}

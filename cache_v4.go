package bigcache

import (
	"math"
	"time"
)

// cacheV4 use shards map
type cacheV4 struct {
	shards    []Cacher
	shardMask uint64
}

func NewCacheV4(capacity int, shards int) Cacher {
	if shards == 0 {
		shards = 128
	}
	if shards&(shards-1) != 0 {
		shards = int(math.Pow(2, math.Ceil(math.Log2(float64(shards)))))
	}
	t := &cacheV4{
		shards:    make([]Cacher, shards),
		shardMask: uint64(shards - 1),
	}
	for i := 0; i < shards; i++ {
		t.shards[i] = NewCacheV2(capacity / shards)
	}
	return t
}

func (t *cacheV4) Set(key string, value []byte, ttl time.Duration) error {
	shardKey := t.getShardKey(key)
	return t.shards[shardKey].Set(key, value, ttl)
}

func (t *cacheV4) Get(key string) ([]byte, error) {
	shardKey := t.getShardKey(key)
	return t.shards[shardKey].Get(key)
}

func (t *cacheV4) TTL(key string) (time.Duration, error) {
	shardKey := t.getShardKey(key)
	return t.shards[shardKey].TTL(key)
}

func (t *cacheV4) Delete(key string) {
	shardKey := t.getShardKey(key)
	t.shards[shardKey].Delete(key)
}

func (t *cacheV4) Len() int {
	i := 0
	for k := range t.shards {
		i += t.shards[k].Len()
	}
	return i
}

func (t *cacheV4) getShardKey(key string) int {
	return int(NewDefaultHasher().Sum64(key) & t.shardMask)
}

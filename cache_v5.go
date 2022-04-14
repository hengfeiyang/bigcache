package bigcache

import (
	"encoding/binary"
	"math"
	"sync"
	"time"
)

const headerLength = 8 + 4 // timestamp:8 + valueLength:4

// cacheV5 use shards map and bytes storage
type cacheV5 struct {
	shards    []*cacheV5ByteStorage
	shardMask uint64
}

type cacheV5ByteStorage struct {
	data     map[uint64]int
	storage  []byte
	lock     sync.RWMutex
	tail     int
	capacity int
}

func NewCacheV5(capacity, maxEntrySize, shards int) Cacher {
	if shards == 0 {
		shards = 128
	}
	if shards&(shards-1) != 0 {
		shards = int(math.Pow(2, math.Ceil(math.Log2(float64(shards)))))
	}
	t := &cacheV5{
		shards:    make([]*cacheV5ByteStorage, shards),
		shardMask: uint64(shards - 1),
	}
	for i := 0; i < shards; i++ {
		t.shards[i] = NewCacheV5ByteStorage(capacity/shards, maxEntrySize)
	}
	return t
}

func (t *cacheV5) Set(key string, value []byte, ttl time.Duration) error {
	keyHash := t.getHashKey(key)
	shardKey := int(keyHash & t.shardMask)
	return t.shards[shardKey].Set(keyHash, value, ttl)
}

func (t *cacheV5) Get(key string) ([]byte, error) {
	keyHash := t.getHashKey(key)
	shardKey := int(keyHash & t.shardMask)
	return t.shards[shardKey].Get(keyHash)
}

func (t *cacheV5) TTL(key string) (time.Duration, error) {
	keyHash := t.getHashKey(key)
	shardKey := int(keyHash & t.shardMask)
	return t.shards[shardKey].TTL(keyHash)
}

func (t *cacheV5) Delete(key string) {
	keyHash := t.getHashKey(key)
	shardKey := int(keyHash & t.shardMask)
	t.shards[shardKey].Delete(keyHash)
}

func (t *cacheV5) Len() int {
	i := 0
	for k := range t.shards {
		i += t.shards[k].Len()
	}
	return i
}

func (t *cacheV5) getHashKey(key string) uint64 {
	return NewDefaultHasher().Sum64(key)

}

func NewCacheV5ByteStorage(capacity, maxEntrySize int) *cacheV5ByteStorage {
	return &cacheV5ByteStorage{
		data:     make(map[uint64]int, capacity),
		storage:  make([]byte, maxEntrySize),
		capacity: maxEntrySize,
	}
}

func (t *cacheV5ByteStorage) Set(keyHash uint64, value []byte, ttl time.Duration) error {
	expires := int64(0)
	if ttl > 0 {
		expires = time.Now().Add(ttl).UnixNano()
	}
	t.lock.Lock()
	needSize := headerLength + len(value)
	if t.capacity-t.tail < needSize {
		// extend storage
		if t.capacity < needSize {
			t.capacity += needSize
		}
		t.capacity = t.capacity * 2
		newStorage := make([]byte, t.capacity)
		copy(newStorage, t.storage)
		t.storage = newStorage
	}
	binary.LittleEndian.PutUint64(t.storage[t.tail:], uint64(expires))      // uint64 -> 8
	binary.LittleEndian.PutUint32(t.storage[t.tail+8:], uint32(len(value))) // uint32 -> 4
	copy(t.storage[t.tail+headerLength:], value)                            // body
	t.data[keyHash] = t.tail
	t.tail += headerLength + len(value)
	t.lock.Unlock()
	return nil
}

func (t *cacheV5ByteStorage) Get(keyHash uint64) ([]byte, error) {
	t.lock.RLock()
	pos, ok := t.data[keyHash]
	if !ok {
		t.lock.RUnlock()
		return nil, ErrNotExist
	}
	header := t.storage[pos : pos+headerLength]
	t.lock.RUnlock()

	expires := int64(binary.LittleEndian.Uint64(header[0:8]))
	bodyLength := int(binary.LittleEndian.Uint32(header[8 : 8+4]))
	if expires > 0 && expires < time.Now().UnixNano() {
		t.Delete(keyHash)
		return nil, ErrNotExist // expired
	}

	t.lock.RLock()
	val := t.storage[pos+int(headerLength) : pos+int(headerLength)+int(bodyLength)]
	t.lock.RUnlock()
	return val, nil
}

func (t *cacheV5ByteStorage) TTL(keyHash uint64) (time.Duration, error) {
	t.lock.RLock()
	pos, ok := t.data[keyHash]
	if !ok {
		t.lock.RUnlock()
		return -1, ErrNotExist
	}
	header := t.storage[pos : pos+headerLength]
	t.lock.RUnlock()

	expires := int64(binary.LittleEndian.Uint64(header[0:8]))
	if expires == 0 {
		return -1, nil
	}
	ttl := time.Duration(expires - time.Now().UnixNano())
	if ttl <= 0 {
		t.Delete(keyHash)
		return -1, ErrNotExist // expired
	}
	return ttl, nil
}

// Delete delete from map, but not release storage
// TODO: release storage
func (t *cacheV5ByteStorage) Delete(keyHash uint64) {
	t.lock.Lock()
	delete(t.data, keyHash)
	t.lock.Unlock()
}

func (t *cacheV5ByteStorage) Len() int {
	t.lock.RLock()
	n := len(t.data)
	t.lock.RUnlock()
	return n
}

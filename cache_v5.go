package bigcache

import (
	"bytes"
	"encoding/binary"
	"math"
	"sync"
	"time"
)

// cacheV5 use shards map and bytes storage
type cacheV5 struct {
	shards  []*cacheV5ByteStorage
	bitMask uint64
}

type cacheV5ByteStorage struct {
	data    map[uint64]int
	storage []byte
	lock    sync.RWMutex
	head    int
	tail    int
	length  int
}

type cacheV5Item struct {
	Expires     uint64
	KeyLength   uint16
	ValueLength uint32
	Key         []byte
	Value       []byte
}

func NewCacheV5(capacity, maxEntrySize, shards int) Cacher {
	if shards == 0 {
		shards = 128
	}
	if shards&(shards-1) != 0 {
		shards = int(math.Pow(2, math.Ceil(math.Log2(float64(shards)))))
	}
	t := &cacheV5{
		shards:  make([]*cacheV5ByteStorage, shards),
		bitMask: uint64(shards - 1),
	}
	for i := 0; i < shards; i++ {
		t.shards[i] = NewCacheV5ByteStorage(capacity/shards, maxEntrySize/shards)
	}
	return t
}

func (t *cacheV5) Set(key string, value []byte, ttl time.Duration) error {
	keyHash := t.getHashKey(key)
	shardKey := int(keyHash & t.bitMask)
	return t.shards[shardKey].Set(keyHash, value, ttl)
}

func (t *cacheV5) Get(key string) ([]byte, error) {
	keyHash := t.getHashKey(key)
	shardKey := int(keyHash & t.bitMask)
	return t.shards[shardKey].Get(keyHash)
}

func (t *cacheV5) TTL(key string) (time.Duration, error) {
	keyHash := t.getHashKey(key)
	shardKey := int(keyHash & t.bitMask)
	return t.shards[shardKey].TTL(keyHash)
}

func (t *cacheV5) Delete(key string) {
	keyHash := t.getHashKey(key)
	shardKey := int(keyHash & t.bitMask)
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
		data:    make(map[uint64]int, capacity),
		storage: make([]byte, maxEntrySize),
	}
}

const headerLength = 8 + 4 // timestamp:8 + valueLength:4

func (t *cacheV5ByteStorage) Set(keyHash uint64, value []byte, ttl time.Duration) error {
	expires := time.Time{}
	if ttl > 0 {
		expires = time.Now().Add(ttl)
	}
	buf := bytes.NewBuffer(nil)
	needSize := headerLength + len(value)
	t.lock.Lock()
	pos := t.tail
	if cap(t.storage)-len(t.storage) < needSize {
		// extend storage
		newStorage := make([]byte, cap(t.storage)+needSize*2)
		copy(newStorage, t.storage)
		t.storage = newStorage
	}
	binary.Write(buf, binary.LittleEndian, expires.Unix())     // int64 -> 8
	binary.Write(buf, binary.LittleEndian, uint32(len(value))) // uint32 -> 4
	copy(t.storage[t.tail:], buf.Bytes())
	copy(t.storage[t.tail+headerLength:], value) // body
	t.data[keyHash] = pos
	t.tail += headerLength + len(value)
	t.length++
	t.lock.Unlock()
	buf.Reset()
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
	buf := bytes.NewBuffer(header)
	var expires int64
	var keyLength uint16
	var bodyLength uint32
	binary.Read(buf, binary.LittleEndian, &expires)
	binary.Read(buf, binary.LittleEndian, &keyLength)
	binary.Read(buf, binary.LittleEndian, &bodyLength)
	buf.Reset()
	if expires > 0 && expires < time.Now().Unix() {
		return nil, ErrNotExist // expired
	}
	t.lock.RLock()
	buf.Write(t.storage[pos+int(headerLength)+int(keyLength) : pos+int(headerLength)+int(keyLength)+int(bodyLength)])
	t.lock.RUnlock()
	return buf.Bytes(), nil
}

func (t *cacheV5ByteStorage) TTL(keyHash uint64) (time.Duration, error) {
	return 0, nil
}

func (t *cacheV5ByteStorage) Delete(keyHash uint64) {

}

func (t *cacheV5ByteStorage) Len() int {
	return t.length
}

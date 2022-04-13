package main

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/hengfeiyang/bigcache"
)

func main() {
	cache := bigcache.NewCacheV4(1024*1024, 128)
	for i := 0; i < 1024*1024; i++ {
		cache.Set(key(i), value(i), 0)
	}
	for i := 0; ; i++ {
		cache.Delete(key(i))
		cache.Set(key(1024*1024+i), value(i), 0)
		time.Sleep(10 * time.Millisecond)
	}
}

func key(i int) string {
	return fmt.Sprintf("key-%d", i)
}

func value(i int) []byte {
	b := bytes.NewBuffer(nil)
	b.WriteString("value-")
	b.Write([]byte(strconv.Itoa(i)))
	return b.Bytes()
}

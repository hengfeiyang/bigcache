package bigcache

import (
	"fmt"
	"testing"
)

func BenchmarkMapSet(b *testing.B) {
	m := make(map[string][]byte, b.N)
	for i := 0; i < b.N; i++ {
		m[key(i)] = value()
	}
}

func BenchmarkCacheV1Set(b *testing.B) {
	c := NewCacheV1(b.N)
	for i := 0; i < b.N; i++ {
		c.Set(key(i), value(), 0)
	}
}

func BenchmarkCacheV2Set(b *testing.B) {
	c := NewCacheV2(b.N)
	for i := 0; i < b.N; i++ {
		c.Set(key(i), value(), 0)
	}
}

func BenchmarkCacheV3Set(b *testing.B) {
	c := NewCacheV3(b.N)
	for i := 0; i < b.N; i++ {
		c.Set(key(i), value(), 0)
	}
}

func BenchmarkCacheV4Set(b *testing.B) {
	c := NewCacheV4(b.N, 0)
	for i := 0; i < b.N; i++ {
		c.Set(key(i), value(), 0)
	}
}

func BenchmarkCacheV5Set(b *testing.B) {
	c := NewCacheV5(b.N, 256, 0)
	for i := 0; i < b.N; i++ {
		c.Set(key(i), value(), 0)
	}
}

func BenchmarkMapGet(b *testing.B) {
	b.StopTimer()
	m := make(map[string][]byte)
	for i := 0; i < b.N; i++ {
		m[key(i)] = value()
	}

	b.StartTimer()
	hitCount := 0
	for i := 0; i < b.N; i++ {
		if m[key(i)] != nil {
			hitCount++
		}
	}
}

func BenchmarkCacherV1Get(b *testing.B) {
	b.StopTimer()
	c := NewCacheV1(b.N)
	for i := 0; i < b.N; i++ {
		c.Set(key(i), value(), 0)
	}

	b.StartTimer()
	hitCount := 0
	for i := 0; i < b.N; i++ {
		if v, _ := c.Get(key(i)); v != nil {
			hitCount++
		}
	}
}

func BenchmarkCacherV2Get(b *testing.B) {
	b.StopTimer()
	c := NewCacheV2(b.N)
	for i := 0; i < b.N; i++ {
		c.Set(key(i), value(), 0)
	}

	b.StartTimer()
	hitCount := 0
	for i := 0; i < b.N; i++ {
		if v, _ := c.Get(key(i)); v != nil {
			hitCount++
		}
	}
}

func BenchmarkCacherV3Get(b *testing.B) {
	b.StopTimer()
	c := NewCacheV3(b.N)
	for i := 0; i < b.N; i++ {
		c.Set(key(i), value(), 0)
	}

	b.StartTimer()
	hitCount := 0
	for i := 0; i < b.N; i++ {
		if v, _ := c.Get(key(i)); v != nil {
			hitCount++
		}
	}
}

func BenchmarkCacherV4Get(b *testing.B) {
	b.StopTimer()
	c := NewCacheV4(b.N, 0)
	for i := 0; i < b.N; i++ {
		c.Set(key(i), value(), 0)
	}

	b.StartTimer()
	hitCount := 0
	for i := 0; i < b.N; i++ {
		if v, _ := c.Get(key(i)); v != nil {
			hitCount++
		}
	}
}

func BenchmarkCacherV2GetParallel(b *testing.B) {
	b.StopTimer()
	c := NewCacheV2(b.N)
	for i := 0; i < b.N; i++ {
		c.Set(key(i), value(), 0)
	}

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			c.Get(key(counter))
			counter = counter + 1
		}
	})
}

func BenchmarkCacherV3GetParallel(b *testing.B) {
	b.StopTimer()
	c := NewCacheV3(b.N)
	for i := 0; i < b.N; i++ {
		c.Set(key(i), value(), 0)
	}

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			c.Get(key(counter))
			counter = counter + 1
		}
	})
}

func BenchmarkCacherV4GetParallel(b *testing.B) {
	b.StopTimer()
	c := NewCacheV4(b.N, 0)
	for i := 0; i < b.N; i++ {
		c.Set(key(i), value(), 0)
	}

	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			c.Get(key(counter))
			counter = counter + 1
		}
	})
}

func key(i int) string {
	return fmt.Sprintf("key-%010d", i)
}

func value() []byte {
	return make([]byte, 100)
}

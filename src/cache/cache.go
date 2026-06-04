package cache

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type CacheStats struct {
	Entradas int    `json:"entradas"`
	Hits     int64  `json:"hits"`
	Misses   int64  `json:"misses"`
	HitRate  string `json:"hit_rate"`
}

type item struct {
	value     any
	expiresAt time.Time
}

type Cache struct {
	items      sync.Map
	hits       int64
	misses     int64
	defaultTTL time.Duration
}

func New(defaultTTL time.Duration) *Cache {
	return &Cache{defaultTTL: defaultTTL}
}

func (c *Cache) Set(key string, value any, ttl time.Duration) {
	if c == nil {
		return
	}

	effectiveTTL := ttl
	if effectiveTTL <= 0 {
		effectiveTTL = c.defaultTTL
	}

	var expiresAt time.Time
	if effectiveTTL > 0 {
		expiresAt = time.Now().Add(effectiveTTL)
	}

	c.items.Store(key, item{value: value, expiresAt: expiresAt})
}

func (c *Cache) Get(key string) (any, bool) {
	if c == nil {
		return nil, false
	}

	loaded, ok := c.items.Load(key)
	if !ok {
		atomic.AddInt64(&c.misses, 1)
		return nil, false
	}

	entry, ok := loaded.(item)
	if !ok {
		c.items.Delete(key)
		atomic.AddInt64(&c.misses, 1)
		return nil, false
	}

	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		c.items.Delete(key)
		atomic.AddInt64(&c.misses, 1)
		return nil, false
	}

	atomic.AddInt64(&c.hits, 1)
	return entry.value, true
}

func (c *Cache) Stats() CacheStats {
	if c == nil {
		return CacheStats{HitRate: "0.0%"}
	}

	stats := CacheStats{
		Hits:    atomic.LoadInt64(&c.hits),
		Misses:  atomic.LoadInt64(&c.misses),
		HitRate: "0.0%",
	}

	c.items.Range(func(_, _ any) bool {
		stats.Entradas++
		return true
	})

	total := stats.Hits + stats.Misses
	if total > 0 {
		hitRate := float64(stats.Hits) / float64(total) * 100
		stats.HitRate = fmt.Sprintf("%.1f%%", hitRate)
	}

	return stats
}

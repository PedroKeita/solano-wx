package cache_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"solano-wx/src/cache"
)

func TestCache_HitBasico(t *testing.T) {
	c := cache.New(time.Minute)

	c.Set("k", "v", time.Minute)
	value, ok := c.Get("k")

	assert.True(t, ok)
	assert.Equal(t, "v", value)
	assert.Equal(t, int64(1), c.Stats().Hits)
}

func TestCache_ExpiraTTL(t *testing.T) {
	c := cache.New(time.Minute)

	c.Set("k", "v", time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	value, ok := c.Get("k")

	assert.False(t, ok)
	assert.Nil(t, value)
	assert.Equal(t, int64(1), c.Stats().Misses)
}

func TestCache_ThreadSafety(t *testing.T) {
	c := cache.New(time.Minute)
	var wg sync.WaitGroup

	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			key := time.Now().UTC().Format(time.RFC3339Nano) + "-" + time.Duration(index).String()
			c.Set(key, index, time.Minute)
			value, ok := c.Get(key)

			assert.True(t, ok)
			assert.Equal(t, index, value)
		}(i)
	}

	wg.Wait()
	stats := c.Stats()
	assert.GreaterOrEqual(t, stats.Entradas, 1)
}

func TestCache_StatsHitRate(t *testing.T) {
	c := cache.New(time.Minute)

	c.Set("a", 1, time.Minute)
	c.Set("b", 2, time.Minute)

	_, _ = c.Get("a")
	_, _ = c.Get("b")
	_, _ = c.Get("a")
	_, _ = c.Get("missing")

	stats := c.Stats()

	assert.Equal(t, int64(3), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
	assert.Equal(t, "75.0%", stats.HitRate)
}

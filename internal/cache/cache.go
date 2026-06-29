package cache

import (
	"maps"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

type Cache[T any] struct {
	impl      *ristretto.Cache[string, T]
	cacheType string
}

func New[T any](costFunc func(T) int64, cacheType string) (*Cache[T], error) {
	impl, err := ristretto.NewCache(&ristretto.Config[string, T]{
		NumCounters: 1e6,
		MaxCost:     1 << 24,
		BufferItems: 64,
		Metrics:     true,
		Cost:        costFunc,
	})
	if err != nil {
		return nil, err
	}
	return &Cache[T]{impl: impl, cacheType: cacheType}, nil
}

func (c *Cache[T]) Get(key string) (T, bool) {
	return c.impl.Get(key)
}

func (c *Cache[T]) SetWithTTL(key string, value T, cost int64,
	ttl time.Duration) bool {
	return c.impl.SetWithTTL(key, value, cost, ttl)
}

func (c *Cache[T]) Clear() {
	c.impl.Clear()
}

func (c *Cache[T]) GetItemCount() int64 {
	return int64(c.impl.Metrics.KeysAdded() - c.impl.Metrics.KeysEvicted())
}

func (c *Cache[T]) Stats() map[string]interface{} {
	metrics := c.impl.Metrics
	memoryUsed := metrics.CostAdded() - metrics.CostEvicted()
	totalRequests := metrics.Hits() + metrics.Misses()
	hitRate := 0.0
	if totalRequests > 0 {
		hitRate = float64(metrics.Hits()) / float64(totalRequests) * 100
	}
	stats := map[string]interface{}{
		"cache_type":     c.cacheType,
		"hits":           metrics.Hits(),
		"misses":         metrics.Misses(),
		"hit_rate":       hitRate,
		"current_items":  c.GetItemCount(),
		"memory_used_kb": float64(memoryUsed) / 1024,
	}
	return stats
}

func (c *Cache[T]) StatsCopy() map[string]interface{} {
	stats := c.Stats()
	out := make(map[string]interface{}, len(stats))
	maps.Copy(out, stats)
	return out
}

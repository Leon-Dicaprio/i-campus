package cache

import (
	"context"
	"sync"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (value []byte, found bool, err error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

type entry struct {
	value     []byte
	expiredAt time.Time
}

type InMemoryCache struct {
	mu   sync.RWMutex
	data map[string]entry
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		data: make(map[string]entry),
	}
}

func (c *InMemoryCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.data[key]
	if !ok {
		return nil, false, nil
	}
	if !e.expiredAt.IsZero() && time.Now().After(e.expiredAt) {
		return nil, false, nil
	}
	return e.value, true, nil
}

func (c *InMemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	e := entry{value: value}
	if ttl > 0 {
		e.expiredAt = time.Now().Add(ttl)
	}
	c.data[key] = e
	return nil
}


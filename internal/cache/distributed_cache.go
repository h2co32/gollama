package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// DistributedCache provides a Redis-based distributed caching mechanism
type DistributedCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewDistributedCache initializes a new DistributedCache with the given Redis address
func NewDistributedCache(redisAddr string) *DistributedCache {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &DistributedCache{
		client: client,
		ctx:    context.Background(),
	}
}

// Set stores a key-value pair in the cache with an expiration duration
func (dc *DistributedCache) Set(key string, data interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := dc.client.Set(dc.ctx, key, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache data: %w", err)
	}
	return nil
}

// Get retrieves a value from the cache by key, returning nil if not found or expired
func (dc *DistributedCache) Get(key string, target interface{}) error {
	jsonData, err := dc.client.Get(dc.ctx, key).Bytes()
	if err == redis.Nil {
		return fmt.Errorf("key not found in cache")
	} else if err != nil {
		return fmt.Errorf("failed to get cache data: %w", err)
	}

	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to unmarshal cache data: %w", err)
	}
	return nil
}

// Delete removes a cached item by key
func (dc *DistributedCache) Delete(key string) error {
	if err := dc.client.Del(dc.ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete cache data: %w", err)
	}
	return nil
}

// Clear flushes all data from the cache
func (dc *DistributedCache) Clear() error {
	if err := dc.client.FlushDB(dc.ctx).Err(); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	return nil
}

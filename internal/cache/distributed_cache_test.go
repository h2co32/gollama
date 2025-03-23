package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDistributedCacheSetGet tests the Set and Get methods of DistributedCache
func TestDistributedCacheSetGet(t *testing.T) {
	// Start a mock Redis server
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer s.Close()

	// Create a new distributed cache with the mock Redis server
	cache := NewDistributedCache(s.Addr())

	// Test data
	type testStruct struct {
		Name  string
		Value int
	}
	key := "test-key"
	value := testStruct{Name: "test", Value: 42}
	ttl := 1 * time.Hour

	// Set the value
	err = cache.Set(key, value, ttl)
	require.NoError(t, err, "Failed to set cache value")

	// Verify the TTL was set correctly in Redis
	redisTTL := s.TTL(key)
	assert.True(t, redisTTL > 0, "Expected TTL to be set")
	assert.True(t, redisTTL <= ttl, "Expected TTL to be less than or equal to the set TTL")

	// Get the value
	var retrieved testStruct
	err = cache.Get(key, &retrieved)
	require.NoError(t, err, "Failed to get cache value")
	assert.Equal(t, value, retrieved, "Retrieved value does not match set value")

	// Test getting a non-existent key
	var nonExistent testStruct
	err = cache.Get("non-existent-key", &nonExistent)
	assert.Error(t, err, "Expected error for non-existent key")
	assert.Contains(t, err.Error(), "key not found", "Expected 'key not found' error")
}

// TestDistributedCacheDelete tests the Delete method of DistributedCache
func TestDistributedCacheDelete(t *testing.T) {
	// Start a mock Redis server
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer s.Close()

	// Create a new distributed cache with the mock Redis server
	cache := NewDistributedCache(s.Addr())

	// Set a value
	key := "delete-key"
	value := "delete-value"
	ttl := 1 * time.Hour

	err = cache.Set(key, value, ttl)
	require.NoError(t, err, "Failed to set cache value")

	// Verify the key exists in Redis
	exists := s.Exists(key)
	assert.True(t, exists, "Expected key to exist in Redis")

	// Delete the key
	err = cache.Delete(key)
	require.NoError(t, err, "Failed to delete cache value")

	// Verify the key no longer exists in Redis
	exists = s.Exists(key)
	assert.False(t, exists, "Expected key to be deleted from Redis")

	// Test deleting a non-existent key (should not error)
	err = cache.Delete("non-existent-key")
	assert.NoError(t, err, "Expected no error when deleting non-existent key")
}

// TestDistributedCacheClear tests the Clear method of DistributedCache
func TestDistributedCacheClear(t *testing.T) {
	// Start a mock Redis server
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer s.Close()

	// Create a new distributed cache with the mock Redis server
	cache := NewDistributedCache(s.Addr())

	// Set multiple values
	keys := []string{"key1", "key2", "key3"}
	ttl := 1 * time.Hour

	for i, key := range keys {
		value := map[string]interface{}{
			"index": i,
			"name":  key,
		}
		err = cache.Set(key, value, ttl)
		require.NoError(t, err, "Failed to set cache value for key '%s'", key)
	}

	// Verify all keys exist in Redis
	for _, key := range keys {
		exists := s.Exists(key)
		assert.True(t, exists, "Expected key '%s' to exist in Redis", key)
	}

	// Clear the cache
	err = cache.Clear()
	require.NoError(t, err, "Failed to clear cache")

	// Verify all keys were removed from Redis
	for _, key := range keys {
		exists := s.Exists(key)
		assert.False(t, exists, "Expected key '%s' to be removed from Redis", key)
	}
}

// TestDistributedCacheExpiration tests the expiration of cached items
func TestDistributedCacheExpiration(t *testing.T) {
	// Start a mock Redis server
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer s.Close()

	// Create a new distributed cache with the mock Redis server
	cache := NewDistributedCache(s.Addr())

	// Set a value with a short TTL
	key := "expiring-key"
	value := "expiring-value"
	ttl := 100 * time.Millisecond

	err = cache.Set(key, value, ttl)
	require.NoError(t, err, "Failed to set cache value")

	// Verify the key exists in Redis
	exists := s.Exists(key)
	assert.True(t, exists, "Expected key to exist in Redis")

	// Fast-forward Redis time to expire the key
	s.FastForward(ttl + 10*time.Millisecond)

	// Verify the key no longer exists in Redis
	exists = s.Exists(key)
	assert.False(t, exists, "Expected key to be expired in Redis")

	// Try to get the expired value
	var retrieved string
	err = cache.Get(key, &retrieved)
	assert.Error(t, err, "Expected error for expired key")
	assert.Contains(t, err.Error(), "key not found", "Expected 'key not found' error")
}

// TestDistributedCacheUnmarshalError tests handling of unmarshal errors
func TestDistributedCacheUnmarshalError(t *testing.T) {
	// Start a mock Redis server
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer s.Close()

	// Create a new distributed cache with the mock Redis server
	cache := NewDistributedCache(s.Addr())

	// Set a value of one type
	key := "type-mismatch"
	value := map[string]interface{}{
		"name": "test",
		"age":  42,
	}
	ttl := 1 * time.Hour

	err = cache.Set(key, value, ttl)
	require.NoError(t, err, "Failed to set cache value")

	// Try to get the value as a different type
	var retrieved string // String can't unmarshal a map
	err = cache.Get(key, &retrieved)
	assert.Error(t, err, "Expected error for type mismatch")
	assert.Contains(t, err.Error(), "unmarshal", "Expected unmarshal error")
}

// TestDistributedCacheRedisError tests handling of Redis connection errors
func TestDistributedCacheRedisError(t *testing.T) {
	// Create a distributed cache with an invalid Redis address
	cache := NewDistributedCache("invalid-address:6379")

	// Try to set a value
	err := cache.Set("key", "value", 1*time.Hour)
	assert.Error(t, err, "Expected error for invalid Redis address")

	// Try to get a value
	var retrieved string
	err = cache.Get("key", &retrieved)
	assert.Error(t, err, "Expected error for invalid Redis address")

	// Try to delete a value
	err = cache.Delete("key")
	assert.Error(t, err, "Expected error for invalid Redis address")

	// Try to clear the cache
	err = cache.Clear()
	assert.Error(t, err, "Expected error for invalid Redis address")
}

// Mock implementation for testing Redis errors
type mockRedisClient struct {
	redis.Client
	ctx context.Context
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd {
	return redis.NewStatusCmd(ctx, "")
}

func (m *mockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	return redis.NewStringCmd(ctx, "")
}

func (m *mockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return redis.NewIntCmd(ctx, "")
}

func (m *mockRedisClient) FlushDB(ctx context.Context) *redis.StatusCmd {
	return redis.NewStatusCmd(ctx, "")
}

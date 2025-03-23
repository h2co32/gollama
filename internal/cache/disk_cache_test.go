package cache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestNewDiskCache(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "disk-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Test creating a new disk cache
	cache, err := NewDiskCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to create disk cache: %v", err)
	}
	
	if cache == nil {
		t.Fatal("Expected NewDiskCache to return a non-nil value")
	}
	
	if cache.directory != tempDir {
		t.Errorf("Expected cache.directory to be '%s', got '%s'", tempDir, cache.directory)
	}
	
	// Test creating a disk cache with a non-existent directory (should create it)
	nonExistentDir := filepath.Join(tempDir, "non-existent")
	cache, err = NewDiskCache(nonExistentDir)
	if err != nil {
		t.Fatalf("Failed to create disk cache with non-existent directory: %v", err)
	}
	
	// Verify the directory was created
	if _, err := os.Stat(nonExistentDir); os.IsNotExist(err) {
		t.Errorf("Expected directory '%s' to be created", nonExistentDir)
	}
}

func TestDiskCacheSetGet(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "disk-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	cache, err := NewDiskCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to create disk cache: %v", err)
	}
	
	// Test setting and getting a value
	key := "test-key"
	value := []byte("test-value")
	ttl := 1 * time.Hour
	
	err = cache.Set(key, value, ttl)
	if err != nil {
		t.Fatalf("Failed to set cache value: %v", err)
	}
	
	// Verify the file was created
	filePath := filepath.Join(tempDir, key+".json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Expected cache file '%s' to be created", filePath)
	}
	
	// Read the file directly to verify its contents
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read cache file: %v", err)
	}
	
	var item CacheItem
	if err := json.Unmarshal(fileData, &item); err != nil {
		t.Fatalf("Failed to unmarshal cache item: %v", err)
	}
	
	if !bytes.Equal(item.Data, value) {
		t.Errorf("Expected item.Data to be '%s', got '%s'", value, item.Data)
	}
	
	// Verify the expiration time is set correctly (within a small margin of error)
	expectedExpiry := time.Now().Add(ttl)
	if item.ExpiresAt.Sub(expectedExpiry) > 1*time.Second {
		t.Errorf("Expected item.ExpiresAt to be close to %v, got %v", expectedExpiry, item.ExpiresAt)
	}
	
	// Test getting the value
	retrievedValue, err := cache.Get(key)
	if err != nil {
		t.Fatalf("Failed to get cache value: %v", err)
	}
	
	if !bytes.Equal(retrievedValue, value) {
		t.Errorf("Expected retrieved value to be '%s', got '%s'", value, retrievedValue)
	}
	
	// Test getting a non-existent key
	nonExistentValue, err := cache.Get("non-existent-key")
	if err != nil {
		t.Fatalf("Expected no error for non-existent key, got %v", err)
	}
	
	if nonExistentValue != nil {
		t.Errorf("Expected nil value for non-existent key, got '%s'", nonExistentValue)
	}
}

func TestDiskCacheExpiration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "disk-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	cache, err := NewDiskCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to create disk cache: %v", err)
	}
	
	// Test setting a value with a short TTL
	key := "expiring-key"
	value := []byte("expiring-value")
	ttl := 10 * time.Millisecond // Very short TTL for testing
	
	err = cache.Set(key, value, ttl)
	if err != nil {
		t.Fatalf("Failed to set cache value: %v", err)
	}
	
	// Wait for the TTL to expire
	time.Sleep(20 * time.Millisecond)
	
	// Try to get the expired value
	retrievedValue, err := cache.Get(key)
	if err != nil {
		t.Fatalf("Expected no error for expired key, got %v", err)
	}
	
	if retrievedValue != nil {
		t.Errorf("Expected nil value for expired key, got '%s'", retrievedValue)
	}
	
	// Verify the file was removed
	filePath := filepath.Join(tempDir, key+".json")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("Expected cache file '%s' to be removed after expiration", filePath)
	}
}

func TestDiskCacheDelete(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "disk-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	cache, err := NewDiskCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to create disk cache: %v", err)
	}
	
	// Set a value
	key := "delete-key"
	value := []byte("delete-value")
	ttl := 1 * time.Hour
	
	err = cache.Set(key, value, ttl)
	if err != nil {
		t.Fatalf("Failed to set cache value: %v", err)
	}
	
	// Verify the file exists
	filePath := filepath.Join(tempDir, key+".json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Expected cache file '%s' to be created", filePath)
	}
	
	// Delete the value
	err = cache.Delete(key)
	if err != nil {
		t.Fatalf("Failed to delete cache value: %v", err)
	}
	
	// Verify the file was removed
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("Expected cache file '%s' to be removed after deletion", filePath)
	}
	
	// Try to get the deleted value
	retrievedValue, err := cache.Get(key)
	if err != nil {
		t.Fatalf("Expected no error for deleted key, got %v", err)
	}
	
	if retrievedValue != nil {
		t.Errorf("Expected nil value for deleted key, got '%s'", retrievedValue)
	}
	
	// Test deleting a non-existent key (should not error)
	err = cache.Delete("non-existent-key")
	if err != nil {
		t.Errorf("Expected no error when deleting non-existent key, got %v", err)
	}
}

func TestDiskCacheClear(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "disk-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	cache, err := NewDiskCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to create disk cache: %v", err)
	}
	
	// Set multiple values
	keys := []string{"key1", "key2", "key3"}
	ttl := 1 * time.Hour
	
	for i, key := range keys {
		value := []byte(fmt.Sprintf("value%d", i+1))
		err = cache.Set(key, value, ttl)
		if err != nil {
			t.Fatalf("Failed to set cache value for key '%s': %v", key, err)
		}
	}
	
	// Verify all files exist
	for _, key := range keys {
		filePath := filepath.Join(tempDir, key+".json")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected cache file '%s' to be created", filePath)
		}
	}
	
	// Clear the cache
	err = cache.Clear()
	if err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}
	
	// Verify all files were removed
	for _, key := range keys {
		filePath := filepath.Join(tempDir, key+".json")
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Errorf("Expected cache file '%s' to be removed after clearing", filePath)
		}
	}
	
	// Verify the directory still exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Expected cache directory '%s' to still exist after clearing", tempDir)
	}
}

func TestDiskCacheConcurrency(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "disk-cache-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	cache, err := NewDiskCache(tempDir)
	if err != nil {
		t.Fatalf("Failed to create disk cache: %v", err)
	}
	
	// Test concurrent access
	const numGoroutines = 10
	const numOperations = 100
	
	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key-%d-%d", id, j)
				value := []byte(fmt.Sprintf("value-%d-%d", id, j))
				ttl := 1 * time.Hour
				
				// Set a value
				err := cache.Set(key, value, ttl)
				if err != nil {
					t.Errorf("Failed to set cache value: %v", err)
					continue
				}
				
				// Get the value
				retrievedValue, err := cache.Get(key)
				if err != nil {
					t.Errorf("Failed to get cache value: %v", err)
					continue
				}
				
				if !bytes.Equal(retrievedValue, value) {
					t.Errorf("Expected retrieved value to be '%s', got '%s'", value, retrievedValue)
				}
				
				// Delete the value
				err = cache.Delete(key)
				if err != nil {
					t.Errorf("Failed to delete cache value: %v", err)
				}
			}
		}(i)
	}
	
	wg.Wait()
}

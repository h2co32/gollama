package cache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DiskCache manages data caching on the local filesystem
type DiskCache struct {
	directory string
	mu        sync.RWMutex
}

// CacheItem represents a single cached item with data and expiration
type CacheItem struct {
	Data      []byte    `json:"data"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewDiskCache initializes a new DiskCache with the specified directory
func NewDiskCache(directory string) (*DiskCache, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	return &DiskCache{directory: directory}, nil
}

// Set stores a key-value pair in the cache with an expiration duration
func (dc *DiskCache) Set(key string, data []byte, ttl time.Duration) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	item := CacheItem{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}

	filePath := filepath.Join(dc.directory, key+".json")
	fileData, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal cache item: %w", err)
	}

	if err := ioutil.WriteFile(filePath, fileData, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}
	return nil
}

// Get retrieves a value from the cache by key, returning nil if expired or not found
func (dc *DiskCache) Get(key string) ([]byte, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	filePath := filepath.Join(dc.directory, key+".json")
	fileData, err := ioutil.ReadFile(filePath)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var item CacheItem
	if err := json.Unmarshal(fileData, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache item: %w", err)
	}

	if time.Now().After(item.ExpiresAt) {
		_ = os.Remove(filePath) // Remove expired item
		return nil, nil
	}

	return item.Data, nil
}

// Delete removes a cached item by key
func (dc *DiskCache) Delete(key string) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	filePath := filepath.Join(dc.directory, key+".json")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cache file: %w", err)
	}
	return nil
}

// Clear removes all cached items
func (dc *DiskCache) Clear() error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	files, err := ioutil.ReadDir(dc.directory)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, file := range files {
		if err := os.Remove(filepath.Join(dc.directory, file.Name())); err != nil {
			return fmt.Errorf("failed to clear cache file: %w", err)
		}
	}
	return nil
}

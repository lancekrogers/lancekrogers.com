package github

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	defaultCacheDir      = "data/cache"
	githubCacheFile      = "github_stats.json"
	defaultCacheExpiry   = 1 * time.Hour
)

// Cache handles caching of GitHub data
type Cache struct {
	mutex    sync.RWMutex
	cacheDir string
	expiry   time.Duration
}

// NewCache creates a new cache instance
func NewCache(cacheDir string, expiry time.Duration) *Cache {
	if cacheDir == "" {
		cacheDir = defaultCacheDir
	}
	if expiry == 0 {
		expiry = defaultCacheExpiry
	}
	
	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		// Log warning but continue - we'll fall back to no caching
		fmt.Printf("Warning: Failed to create cache directory %s: %v\n", cacheDir, err)
	}
	
	return &Cache{
		cacheDir: cacheDir,
		expiry:   expiry,
	}
}

// Get retrieves cached GitHub stats if they exist and are not expired
func (c *Cache) Get() (*GitHubStats, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	filePath := filepath.Join(c.cacheDir, githubCacheFile)
	
	// Check if cache file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, false
	}
	
	// Read cache file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, false
	}
	
	// Parse cache entry
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}
	
	// Check if cache has expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	
	return entry.Data, true
}

// Set stores GitHub stats in cache
func (c *Cache) Set(stats *GitHubStats) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	now := time.Now()
	entry := CacheEntry{
		Data:      stats,
		Timestamp: now,
		ExpiresAt: now.Add(c.expiry),
	}
	
	// Marshal cache entry
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}
	
	// Write to cache file
	filePath := filepath.Join(c.cacheDir, githubCacheFile)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}
	
	return nil
}

// Clear removes the cached data
func (c *Cache) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	filePath := filepath.Join(c.cacheDir, githubCacheFile)
	
	// Check if file exists before attempting to remove
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to clear
	}
	
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove cache file: %w", err)
	}
	
	return nil
}

// IsExpired checks if the cache has expired
func (c *Cache) IsExpired() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	filePath := filepath.Join(c.cacheDir, githubCacheFile)
	
	// Check if cache file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return true
	}
	
	// Read cache file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return true
	}
	
	// Parse cache entry
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return true
	}
	
	// Check if cache has expired
	return time.Now().After(entry.ExpiresAt)
}
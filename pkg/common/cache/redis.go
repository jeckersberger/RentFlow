package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// RedisCache provides a caching interface
// In production, use: github.com/redis/go-redis/v9
type RedisCache struct {
	addr     string
	password string
	db       int
	mu       sync.RWMutex
	cache    map[string]*cacheEntry
}

// cacheEntry represents a cached value with expiration
type cacheEntry struct {
	value      interface{}
	expiresAt  *time.Time
	lastAccess time.Time
}

// NewRedisCache creates a new Redis cache wrapper
// Note: This is a simple in-memory implementation
// In production, connect to an actual Redis server using github.com/redis/go-redis
func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	if addr == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	return &RedisCache{
		addr:     addr,
		password: password,
		db:       db,
		cache:    make(map[string]*cacheEntry),
	}, nil
}

// Get retrieves a value from the cache
func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return "", ErrKeyNotFound
	}

	// Check if entry has expired
	if entry.expiresAt != nil && time.Now().After(*entry.expiresAt) {
		c.mu.RUnlock()
		c.mu.Lock()
		delete(c.cache, key)
		c.mu.Unlock()
		c.mu.RLock()
		return "", ErrKeyNotFound
	}

	// Update last access time
	entry.lastAccess = time.Now()

	// Convert to string
	switch v := entry.value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to marshal value: %w", err)
		}
		return string(data), nil
	}
}

// Set stores a value in the cache with optional TTL
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt *time.Time
	if ttl > 0 {
		t := time.Now().Add(ttl)
		expiresAt = &t
	}

	c.cache[key] = &cacheEntry{
		value:      value,
		expiresAt:  expiresAt,
		lastAccess: time.Now(),
	}

	return nil
}

// Delete removes a value from the cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, key)
	return nil
}

// Exists checks if a key exists in the cache
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return false, nil
	}

	// Check if entry has expired
	if entry.expiresAt != nil && time.Now().After(*entry.expiresAt) {
		c.mu.RUnlock()
		c.mu.Lock()
		delete(c.cache, key)
		c.mu.Unlock()
		return false, nil
	}

	return true, nil
}

// GetInt retrieves an integer value from the cache
func (c *RedisCache) GetInt(ctx context.Context, key string) (int, error) {
	val, err := c.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	var result int
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return 0, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return result, nil
}

// GetString retrieves a string value from the cache
func (c *RedisCache) GetString(ctx context.Context, key string) (string, error) {
	return c.Get(ctx, key)
}

// SetString stores a string value in the cache
func (c *RedisCache) SetString(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.Set(ctx, key, value, ttl)
}

// IncrementInt increments an integer value in the cache
func (c *RedisCache) IncrementInt(ctx context.Context, key string, delta int) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.cache[key]
	if !exists {
		c.cache[key] = &cacheEntry{
			value:      delta,
			lastAccess: time.Now(),
		}
		return delta, nil
	}

	// Convert to int
	var current int
	switch v := entry.value.(type) {
	case int:
		current = v
	case float64:
		current = int(v)
	case string:
		if _, err := fmt.Sscanf(v, "%d", &current); err != nil {
			return 0, fmt.Errorf("failed to parse value as int: %w", err)
		}
	default:
		return 0, fmt.Errorf("value is not an integer")
	}

	newValue := current + delta
	entry.value = newValue
	entry.lastAccess = time.Now()

	return newValue, nil
}

// AppendToList appends a value to a list in the cache
func (c *RedisCache) AppendToList(ctx context.Context, key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.cache[key]
	if !exists {
		c.cache[key] = &cacheEntry{
			value:      []interface{}{value},
			lastAccess: time.Now(),
		}
		return nil
	}

	// Ensure value is a list
	list, ok := entry.value.([]interface{})
	if !ok {
		return fmt.Errorf("value is not a list")
	}

	entry.value = append(list, value)
	entry.lastAccess = time.Now()

	return nil
}

// GetList retrieves a list from the cache
func (c *RedisCache) GetList(ctx context.Context, key string) ([]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil, ErrKeyNotFound
	}

	list, ok := entry.value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("value is not a list")
	}

	entry.lastAccess = time.Now()
	return list, nil
}

// Clear removes all entries from the cache
func (c *RedisCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*cacheEntry)
	return nil
}

// Health checks the health of the cache
func (c *RedisCache) Health(ctx context.Context) error {
	// Simple health check
	_, err := c.Get(ctx, "health_check")
	if err == ErrKeyNotFound {
		return nil
	}
	return err
}

// Stats returns cache statistics
type CacheStats struct {
	Size      int
	Entries   int
	BytesUsed int64
}

// GetStats returns cache statistics
func (c *RedisCache) GetStats(ctx context.Context) CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := CacheStats{
		Size:    len(c.cache),
		Entries: len(c.cache),
	}

	return stats
}

// Close closes the cache connection
func (c *RedisCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*cacheEntry)
	return nil
}

// Cache interface for abstracting the cache implementation
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
	Health(ctx context.Context) error
	Close() error
}

// Ensure RedisCache implements Cache interface
var _ Cache = (*RedisCache)(nil)

// Cache errors
var (
	ErrKeyNotFound = fmt.Errorf("key not found")
)

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the Redis client and provides a caching interface
type RedisClient struct {
	client *redis.Client
	mu     sync.RWMutex
	fallback *InMemoryCache // fallback for testing/dev when Redis is unavailable
}

// InMemoryCache is a fallback in-memory cache for testing/development
type InMemoryCache struct {
	mu    sync.RWMutex
	cache map[string]*cacheEntry
}

// cacheEntry represents a cached value with expiration
type cacheEntry struct {
	value      interface{}
	expiresAt  *time.Time
	lastAccess time.Time
}

// NewRedisClient creates a new Redis client
func NewRedisClient(addr, password string, db int) (*RedisClient, error) {
	if addr == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		// Return error but still create the client for potential retry
		return &RedisClient{
			client:   client,
			fallback: NewInMemoryCache(),
		}, fmt.Errorf("failed to connect to Redis at %s: %w", addr, err)
	}

	return &RedisClient{
		client:   client,
		fallback: NewInMemoryCache(),
	}, nil
}

// NewInMemoryCache creates a new in-memory cache for fallback/testing
func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		cache: make(map[string]*cacheEntry),
	}
}

// NewCacheFromEnv creates a Redis client from environment variables
// Falls back to InMemoryCache if Redis is not configured
func NewCacheFromEnv() *RedisClient {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	
	password := os.Getenv("REDIS_PASSWORD")
	
	// Try to create a real Redis client
	client, err := NewRedisClient(redisURL, password, 0)
	if err != nil {
		// Log but continue with fallback
		fmt.Fprintf(os.Stderr, "Warning: failed to connect to Redis, using in-memory cache: %v\n", err)
		return &RedisClient{
			client:   nil,
			fallback: NewInMemoryCache(),
		}
	}
	
	return client
}

// Get retrieves a value from the cache
func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	if c.client != nil {
		val, err := c.client.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				return "", ErrKeyNotFound
			}
			// Fall back to in-memory cache on error
			return c.fallback.Get(ctx, key)
		}
		return val, nil
	}
	
	// Use fallback
	return c.fallback.Get(ctx, key)
}

// Set stores a value in the cache with optional TTL
func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Convert value to string/bytes for Redis
	var strVal string
	switch v := value.(type) {
	case string:
		strVal = v
	case []byte:
		strVal = string(v)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		strVal = string(data)
	}

	if c.client != nil {
		if err := c.client.Set(ctx, key, strVal, ttl).Err(); err != nil {
			// Fall back to in-memory cache on error
			return c.fallback.Set(ctx, key, value, ttl)
		}
		return nil
	}
	
	// Use fallback
	return c.fallback.Set(ctx, key, value, ttl)
}

// Delete removes a value from the cache
func (c *RedisClient) Delete(ctx context.Context, key string) error {
	if c.client != nil {
		if err := c.client.Del(ctx, key).Err(); err != nil {
			// Still try fallback
			_ = c.fallback.Delete(ctx, key)
			return err
		}
	}
	
	// Also remove from fallback
	return c.fallback.Delete(ctx, key)
}

// Exists checks if a key exists in the cache
func (c *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	if c.client != nil {
		exists, err := c.client.Exists(ctx, key).Result()
		if err != nil {
			// Fall back to in-memory cache on error
			return c.fallback.Exists(ctx, key)
		}
		return exists > 0, nil
	}
	
	// Use fallback
	return c.fallback.Exists(ctx, key)
}

// GetInt retrieves an integer value from the cache
func (c *RedisClient) GetInt(ctx context.Context, key string) (int, error) {
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
func (c *RedisClient) GetString(ctx context.Context, key string) (string, error) {
	return c.Get(ctx, key)
}

// SetString stores a string value in the cache
func (c *RedisClient) SetString(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.Set(ctx, key, value, ttl)
}

// SetJSON stores a JSON-serializable value in the cache
func (c *RedisClient) SetJSON(ctx context.Context, key string, v interface{}, ttl time.Duration) error {
	return c.Set(ctx, key, v, ttl)
}

// GetJSON retrieves and unmarshals a JSON value from the cache
func (c *RedisClient) GetJSON(ctx context.Context, key string, v interface{}) error {
	val, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), v)
}

// IncrementInt increments an integer value in the cache
func (c *RedisClient) IncrementInt(ctx context.Context, key string, delta int) (int, error) {
	if c.client != nil {
		result, err := c.client.IncrBy(ctx, key, int64(delta)).Result()
		if err != nil {
			// Fall back to in-memory cache on error
			return c.fallback.IncrementInt(ctx, key, delta)
		}
		return int(result), nil
	}
	
	// Use fallback
	return c.fallback.IncrementInt(ctx, key, delta)
}

// AppendToList appends a value to a list in the cache
func (c *RedisClient) AppendToList(ctx context.Context, key string, value interface{}) error {
	// Convert value to string
	var strVal string
	switch v := value.(type) {
	case string:
		strVal = v
	case []byte:
		strVal = string(v)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		strVal = string(data)
	}

	if c.client != nil {
		if err := c.client.RPush(ctx, key, strVal).Err(); err != nil {
			// Fall back to in-memory cache on error
			return c.fallback.AppendToList(ctx, key, value)
		}
		return nil
	}
	
	// Use fallback
	return c.fallback.AppendToList(ctx, key, value)
}

// GetList retrieves a list from the cache
func (c *RedisClient) GetList(ctx context.Context, key string) ([]interface{}, error) {
	if c.client != nil {
		result, err := c.client.LRange(ctx, key, 0, -1).Result()
		if err != nil {
			// Fall back to in-memory cache on error
			return c.fallback.GetList(ctx, key)
		}
		
		// Convert strings to interface{}
		list := make([]interface{}, len(result))
		for i, v := range result {
			list[i] = v
		}
		return list, nil
	}
	
	// Use fallback
	return c.fallback.GetList(ctx, key)
}

// Clear removes all entries from the cache
func (c *RedisClient) Clear(ctx context.Context) error {
	if c.client != nil {
		if err := c.client.FlushDB(ctx).Err(); err != nil {
			// Still try fallback
			_ = c.fallback.Clear(ctx)
			return err
		}
	}
	
	// Also clear fallback
	return c.fallback.Clear(ctx)
}

// Health checks the health of the cache
func (c *RedisClient) Health(ctx context.Context) error {
	if c.client != nil {
		return c.client.Ping(ctx).Err()
	}
	// Fallback is always healthy
	return nil
}

// CacheStats represents cache statistics
type CacheStats struct {
	Size      int
	Entries   int
	BytesUsed int64
}

// GetStats returns cache statistics
func (c *RedisClient) GetStats(ctx context.Context) CacheStats {
	// For Redis, we can get some info, but for simplicity return basic stats
	stats := CacheStats{
		Size:      0,
		Entries:   0,
		BytesUsed: 0,
	}
	
	if c.client != nil {
		info := c.client.Info(ctx, "stats")
		if info.Err() == nil {
			// Could parse the stats here, but for now keep it simple
		}
	}
	
	return stats
}

// Close closes the cache connection
func (c *RedisClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return c.fallback.Close()
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

// Ensure RedisClient implements Cache interface
var _ Cache = (*RedisClient)(nil)

// Cache errors
var (
	ErrKeyNotFound = fmt.Errorf("key not found")
)

// --- InMemoryCache Implementation (Fallback) ---

// Get retrieves a value from the in-memory cache
func (im *InMemoryCache) Get(ctx context.Context, key string) (string, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	entry, exists := im.cache[key]
	if !exists {
		return "", ErrKeyNotFound
	}

	// Check if entry has expired
	if entry.expiresAt != nil && time.Now().After(*entry.expiresAt) {
		im.mu.RUnlock()
		im.mu.Lock()
		delete(im.cache, key)
		im.mu.Unlock()
		im.mu.RLock()
		return "", ErrKeyNotFound
	}

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

// Set stores a value in the in-memory cache with optional TTL
func (im *InMemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	var expiresAt *time.Time
	if ttl > 0 {
		t := time.Now().Add(ttl)
		expiresAt = &t
	}

	im.cache[key] = &cacheEntry{
		value:      value,
		expiresAt:  expiresAt,
		lastAccess: time.Now(),
	}

	return nil
}

// Delete removes a value from the in-memory cache
func (im *InMemoryCache) Delete(ctx context.Context, key string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	delete(im.cache, key)
	return nil
}

// Exists checks if a key exists in the in-memory cache
func (im *InMemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	entry, exists := im.cache[key]
	if !exists {
		return false, nil
	}

	// Check if entry has expired
	if entry.expiresAt != nil && time.Now().After(*entry.expiresAt) {
		im.mu.RUnlock()
		im.mu.Lock()
		delete(im.cache, key)
		im.mu.Unlock()
		return false, nil
	}

	return true, nil
}

// IncrementInt increments an integer value in the in-memory cache
func (im *InMemoryCache) IncrementInt(ctx context.Context, key string, delta int) (int, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	entry, exists := im.cache[key]
	if !exists {
		im.cache[key] = &cacheEntry{
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

// AppendToList appends a value to a list in the in-memory cache
func (im *InMemoryCache) AppendToList(ctx context.Context, key string, value interface{}) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	entry, exists := im.cache[key]
	if !exists {
		im.cache[key] = &cacheEntry{
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

// GetList retrieves a list from the in-memory cache
func (im *InMemoryCache) GetList(ctx context.Context, key string) ([]interface{}, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	entry, exists := im.cache[key]
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

// Clear removes all entries from the in-memory cache
func (im *InMemoryCache) Clear(ctx context.Context) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.cache = make(map[string]*cacheEntry)
	return nil
}

// Health checks the health of the in-memory cache
func (im *InMemoryCache) Health(ctx context.Context) error {
	// In-memory cache is always healthy
	return nil
}

// Close closes the in-memory cache
func (im *InMemoryCache) Close() error {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.cache = make(map[string]*cacheEntry)
	return nil
}

// Backward compatibility alias for tests
// NewRedisCache is an alias for NewRedisClient for backward compatibility
func NewRedisCache(addr, password string, db int) (*RedisClient, error) {
	return NewRedisClient(addr, password, db)
}

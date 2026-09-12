package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps Redis and optionally carries an in-memory fallback.
type RedisClient struct {
	client   *redis.Client
	fallback *InMemoryCache
}

// InMemoryCache is used explicitly for tests/development fallback paths.
type InMemoryCache struct {
	mu    sync.RWMutex
	cache map[string]*cacheEntry
}

type cacheEntry struct {
	value      interface{}
	expiresAt  *time.Time
	lastAccess time.Time
}

// NewRedisClient creates a Redis-backed cache and verifies connectivity.
// A connection failure is returned to the caller; production code must decide
// explicitly whether falling back is acceptable.
func NewRedisClient(addr, password string, db int) (*RedisClient, error) {
	if addr == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	client := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wrapper := &RedisClient{client: client, fallback: NewInMemoryCache()}
	if err := client.Ping(ctx).Err(); err != nil {
		return wrapper, fmt.Errorf("failed to connect to Redis at %s: %w", addr, err)
	}
	return wrapper, nil
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{cache: make(map[string]*cacheEntry)}
}

// NewCacheFromEnv is the explicitly best-effort constructor used by code paths
// that are allowed to continue with process-local cache state.
func NewCacheFromEnv() *RedisClient {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	client, err := NewRedisClient(redisURL, os.Getenv("REDIS_PASSWORD"), 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to connect to Redis, using in-memory cache: %v\n", err)
		return &RedisClient{fallback: NewInMemoryCache()}
	}
	return client
}

func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	if c.client == nil {
		return c.fallback.Get(ctx, key)
	}
	val, err := c.client.Get(ctx, key).Result()
	if err == nil {
		return val, nil
	}
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return c.fallback.Get(ctx, key)
}

func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
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

	if c.client == nil {
		return c.fallback.Set(ctx, key, value, ttl)
	}
	if err := c.client.Set(ctx, key, strVal, ttl).Err(); err != nil {
		return c.fallback.Set(ctx, key, value, ttl)
	}
	return nil
}

func (c *RedisClient) Delete(ctx context.Context, key string) error {
	if c.client == nil {
		return c.fallback.Delete(ctx, key)
	}
	if err := c.client.Del(ctx, key).Err(); err != nil {
		_ = c.fallback.Delete(ctx, key)
		return err
	}
	return c.fallback.Delete(ctx, key)
}

func (c *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	if c.client == nil {
		return c.fallback.Exists(ctx, key)
	}
	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return c.fallback.Exists(ctx, key)
	}
	return exists > 0, nil
}

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

func (c *RedisClient) GetString(ctx context.Context, key string) (string, error) {
	return c.Get(ctx, key)
}

func (c *RedisClient) SetString(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.Set(ctx, key, value, ttl)
}

func (c *RedisClient) SetJSON(ctx context.Context, key string, v interface{}, ttl time.Duration) error {
	return c.Set(ctx, key, v, ttl)
}

func (c *RedisClient) GetJSON(ctx context.Context, key string, v interface{}) error {
	val, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), v)
}

func (c *RedisClient) IncrementInt(ctx context.Context, key string, delta int) (int, error) {
	if c.client == nil {
		return c.fallback.IncrementInt(ctx, key, delta)
	}
	result, err := c.client.IncrBy(ctx, key, int64(delta)).Result()
	if err != nil {
		return c.fallback.IncrementInt(ctx, key, delta)
	}
	return int(result), nil
}

func (c *RedisClient) AppendToList(ctx context.Context, key string, value interface{}) error {
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

	if c.client == nil {
		return c.fallback.AppendToList(ctx, key, value)
	}
	if err := c.client.RPush(ctx, key, strVal).Err(); err != nil {
		return c.fallback.AppendToList(ctx, key, value)
	}
	return nil
}

func (c *RedisClient) GetList(ctx context.Context, key string) ([]interface{}, error) {
	if c.client == nil {
		return c.fallback.GetList(ctx, key)
	}
	result, err := c.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return c.fallback.GetList(ctx, key)
	}
	list := make([]interface{}, len(result))
	for i, v := range result {
		list[i] = v
	}
	return list, nil
}

func (c *RedisClient) Clear(ctx context.Context) error {
	if c.client == nil {
		return c.fallback.Clear(ctx)
	}
	if err := c.client.FlushDB(ctx).Err(); err != nil {
		_ = c.fallback.Clear(ctx)
		return err
	}
	return c.fallback.Clear(ctx)
}

func (c *RedisClient) Health(ctx context.Context) error {
	if c.client == nil {
		return c.fallback.Health(ctx)
	}
	return c.client.Ping(ctx).Err()
}

type CacheStats struct {
	Size      int
	Entries   int
	BytesUsed int64
}

func (c *RedisClient) GetStats(ctx context.Context) CacheStats {
	if c.client == nil {
		return c.fallback.GetStats()
	}
	count, err := c.client.DBSize(ctx).Result()
	if err != nil {
		return c.fallback.GetStats()
	}
	return CacheStats{Size: int(count), Entries: int(count)}
}

func (c *RedisClient) Close() error {
	var redisErr error
	if c.client != nil {
		redisErr = c.client.Close()
	}
	_ = c.fallback.Close()
	return redisErr
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
	Health(ctx context.Context) error
	Close() error
}

var _ Cache = (*RedisClient)(nil)

var ErrKeyNotFound = fmt.Errorf("key not found")

func (im *InMemoryCache) Get(ctx context.Context, key string) (string, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	entry, exists := im.cache[key]
	if !exists {
		return "", ErrKeyNotFound
	}
	if entry.expiresAt != nil && time.Now().After(*entry.expiresAt) {
		delete(im.cache, key)
		return "", ErrKeyNotFound
	}
	entry.lastAccess = time.Now()

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

func (im *InMemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	var expiresAt *time.Time
	if ttl > 0 {
		expires := time.Now().Add(ttl)
		expiresAt = &expires
	}
	im.cache[key] = &cacheEntry{value: value, expiresAt: expiresAt, lastAccess: time.Now()}
	return nil
}

func (im *InMemoryCache) Delete(ctx context.Context, key string) error {
	im.mu.Lock()
	defer im.mu.Unlock()
	delete(im.cache, key)
	return nil
}

func (im *InMemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	entry, exists := im.cache[key]
	if !exists {
		return false, nil
	}
	if entry.expiresAt != nil && time.Now().After(*entry.expiresAt) {
		delete(im.cache, key)
		return false, nil
	}
	entry.lastAccess = time.Now()
	return true, nil
}

func (im *InMemoryCache) IncrementInt(ctx context.Context, key string, delta int) (int, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	entry, exists := im.cache[key]
	if !exists {
		im.cache[key] = &cacheEntry{value: delta, lastAccess: time.Now()}
		return delta, nil
	}

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

	current += delta
	entry.value = current
	entry.lastAccess = time.Now()
	return current, nil
}

func (im *InMemoryCache) AppendToList(ctx context.Context, key string, value interface{}) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	entry, exists := im.cache[key]
	if !exists {
		im.cache[key] = &cacheEntry{value: []interface{}{value}, lastAccess: time.Now()}
		return nil
	}
	list, ok := entry.value.([]interface{})
	if !ok {
		return fmt.Errorf("value is not a list")
	}
	entry.value = append(list, value)
	entry.lastAccess = time.Now()
	return nil
}

func (im *InMemoryCache) GetList(ctx context.Context, key string) ([]interface{}, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	entry, exists := im.cache[key]
	if !exists {
		return nil, ErrKeyNotFound
	}
	if entry.expiresAt != nil && time.Now().After(*entry.expiresAt) {
		delete(im.cache, key)
		return nil, ErrKeyNotFound
	}
	list, ok := entry.value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("value is not a list")
	}
	entry.lastAccess = time.Now()
	return list, nil
}

func (im *InMemoryCache) Clear(ctx context.Context) error {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.cache = make(map[string]*cacheEntry)
	return nil
}

func (im *InMemoryCache) Health(ctx context.Context) error { return nil }

func (im *InMemoryCache) GetStats() CacheStats {
	im.mu.RLock()
	defer im.mu.RUnlock()
	entries := len(im.cache)
	return CacheStats{Size: entries, Entries: entries}
}

func (im *InMemoryCache) Close() error {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.cache = make(map[string]*cacheEntry)
	return nil
}

// NewRedisCache retains the historical constructor contract. Production remains
// strict by default; tests/development may opt into process-local fallback with
// CACHE_ALLOW_IN_MEMORY_FALLBACK=true.
func NewRedisCache(addr, password string, db int) (*RedisClient, error) {
	client, err := NewRedisClient(addr, password, db)
	if err == nil {
		return client, nil
	}
	allowFallback, parseErr := strconv.ParseBool(os.Getenv("CACHE_ALLOW_IN_MEMORY_FALLBACK"))
	if parseErr == nil && allowFallback && client != nil {
		if client.client != nil {
			_ = client.client.Close()
		}
		client.client = nil
		return client, nil
	}
	return client, err
}

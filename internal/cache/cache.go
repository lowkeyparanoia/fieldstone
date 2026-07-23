// Package cache provides caching functionality for Fieldstone.
// Supports Redis and in-memory backends with TTL and invalidation.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"strings"
	"sync"
)

// Cache defines the caching interface
type Cache interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) (string, error)

	// GetJSON retrieves and unmarshals JSON data
	GetJSON(ctx context.Context, key string, dest interface{}) error

	// Set stores a value with TTL
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// SetJSON marshals and stores data
	SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a key
	Delete(ctx context.Context, key string) error

	// DeletePattern removes keys matching pattern
	DeletePattern(ctx context.Context, pattern string) error

	// Increment atomically increments a counter
	Increment(ctx context.Context, key string, delta int64) (int64, error)

	// Expire sets/updates TTL
	Expire(ctx context.Context, key string, ttl time.Duration) error

	// TTL returns remaining time to live
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Exists checks if key exists
	Exists(ctx context.Context, key string) (bool, error)

	// Flush clears all cache
	Flush(ctx context.Context) error

	// Close closes cache connection
	Close() error

	// Stats returns cache statistics
	Stats(ctx context.Context) (Stats, error)
}

// Stats represents cache statistics
type Stats struct {
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	KeysCount  int64 `json:"keys_count"`
	MemoryUsed int64 `json:"memory_used"`
}

// Config holds cache configuration
type Config struct {
	Type       string        // "redis", "memory"
	RedisAddr  string        // Redis address (e.g., "localhost:6379")
	RedisPass  string        // Redis password
	RedisDB    int           // Redis database number
	DefaultTTL time.Duration // Default TTL for cached items
	Prefix     string        // Key prefix (for multi-tenancy)
}

// New creates a new cache instance
func New(cfg Config) (Cache, error) {
	switch cfg.Type {
	case "redis":
		return NewRedisCache(cfg)
	case "memory":
		return NewMemoryCache(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported cache type: %s", cfg.Type)
	}
}

// RedisCache implements Cache using Redis
type RedisCache struct {
	client     *redis.Client
	defaultTTL time.Duration
	prefix     string
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(cfg Config) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info().
		Str("addr", cfg.RedisAddr).
		Int("db", cfg.RedisDB).
		Msg("Redis cache connected")

	return &RedisCache{
		client:     client,
		defaultTTL: cfg.DefaultTTL,
		prefix:     cfg.Prefix,
	}, nil
}

func (c *RedisCache) key(k string) string {
	if c.prefix != "" {
		return fmt.Sprintf("%s:%s", c.prefix, k)
	}
	return k
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, c.key(key)).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found: %s", key)
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (c *RedisCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.defaultTTL
	}
	return c.client.Set(ctx, c.key(key), value, ttl).Err()
}

func (c *RedisCache) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, string(data), ttl)
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.key(key)).Err()
}

func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := c.client.Keys(ctx, c.key(pattern)).Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

func (c *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return c.client.IncrBy(ctx, c.key(key), delta).Result()
}

func (c *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(ctx, c.key(key), ttl).Err()
}

func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, c.key(key)).Result()
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, c.key(key)).Result()
	return n > 0, err
}

func (c *RedisCache) Flush(ctx context.Context) error {
	if c.prefix != "" {
		// Only delete keys with our prefix
		keys, err := c.client.Keys(ctx, c.key("*")).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			return c.client.Del(ctx, keys...).Err()
		}
		return nil
	}
	return c.client.FlushDB(ctx).Err()
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

func (c *RedisCache) Stats(ctx context.Context) (Stats, error) {
	_ = c.client.Info(ctx, "stats").Val()
	// Parse Redis INFO stats (simplified)
	// In production, parse the info string properly
	return Stats{}, nil
}

// MemoryCache implements in-memory cache for testing/development.
//
// The map was previously shared across every request goroutine with no
// synchronisation whatsoever. That is not a data race you can survive: Go
// detects concurrent map writes in the runtime and calls fatal(), which is not
// recoverable and takes the whole process down. A mutex is not optional here.
type MemoryCache struct {
	mu         sync.RWMutex
	data       map[string]cacheItem
	defaultTTL time.Duration
	prefix     string
	stats      Stats
}

type cacheItem struct {
	value     string
	expiresAt time.Time
}

// NewMemoryCache creates an in-memory cache
func NewMemoryCache(cfg Config) *MemoryCache {
	return &MemoryCache{
		data:       make(map[string]cacheItem),
		defaultTTL: cfg.DefaultTTL,
		prefix:     cfg.Prefix,
	}
}

func (c *MemoryCache) key(k string) string {
	if c.prefix != "" {
		return fmt.Sprintf("%s:%s", c.prefix, k)
	}
	return k
}

func (c *MemoryCache) Get(ctx context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	k := c.key(key)
	item, ok := c.data[k]
	if !ok {
		c.stats.Misses++
		return "", fmt.Errorf("key not found: %s", key)
	}

	if time.Now().After(item.expiresAt) {
		delete(c.data, k)
		c.stats.Misses++
		return "", fmt.Errorf("key expired: %s", key)
	}

	c.stats.Hits++
	return item.value, nil
}

func (c *MemoryCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

func (c *MemoryCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ttl == 0 {
		ttl = c.defaultTTL
	}
	c.data[c.key(key)] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (c *MemoryCache) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, string(data), ttl)
}

func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, c.key(key))
	return nil
}

// DeletePattern removes every key beginning with pattern.
//
// The previous implementation guarded on
//
//	len(k) >= len(prefix)+len(pattern)
//
// but then sliced to len(prefix)+1+len(pattern), one byte further, so a key
// whose length exactly matched the guard panicked:
//
//	slice bounds out of range [:62] with length 61
//
// That fired on every record creation, because creating a record invalidates
// the collection cache. strings.HasPrefix removes the arithmetic entirely.
func (c *MemoryCache) DeletePattern(ctx context.Context, pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	full := c.key(pattern)
	for k := range c.data {
		if strings.HasPrefix(k, full) {
			delete(c.data, k) // deleting during range is safe in Go
		}
	}
	return nil
}

func (c *MemoryCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	k := c.key(key)
	val, err := c.Get(ctx, key)
	if err != nil {
		c.data[k] = cacheItem{value: fmt.Sprintf("%d", delta), expiresAt: time.Now().Add(c.defaultTTL)}
		return delta, nil
	}

	var current int64
	fmt.Sscanf(val, "%d", &current)
	current += delta
	c.data[k] = cacheItem{value: fmt.Sprintf("%d", current), expiresAt: time.Now().Add(c.defaultTTL)}
	return current, nil
}

func (c *MemoryCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	k := c.key(key)
	item, ok := c.data[k]
	if !ok {
		return fmt.Errorf("key not found")
	}
	item.expiresAt = time.Now().Add(ttl)
	c.data[k] = item
	return nil
}

func (c *MemoryCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	k := c.key(key)
	item, ok := c.data[k]
	if !ok {
		return 0, fmt.Errorf("key not found")
	}
	return time.Until(item.expiresAt), nil
}

func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, err := c.Get(ctx, key)
	return err == nil, nil
}

func (c *MemoryCache) Flush(ctx context.Context) error {
	if c.prefix != "" {
		for k := range c.data {
			if len(k) > len(c.prefix) && k[:len(c.prefix)] == c.prefix {
				delete(c.data, k)
			}
		}
	} else {
		c.data = make(map[string]cacheItem)
	}
	return nil
}

func (c *MemoryCache) Close() error {
	return nil
}

func (c *MemoryCache) Stats(ctx context.Context) (Stats, error) {
	c.stats.KeysCount = int64(len(c.data))
	return c.stats, nil
}

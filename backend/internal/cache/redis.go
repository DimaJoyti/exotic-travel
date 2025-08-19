package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis configuration
type Config struct {
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Cache wraps Redis client with additional functionality
type Cache struct {
	client *redis.Client
	config Config
}

// NewCache creates a new Redis cache client with optimized settings
func NewCache(config Config) (*Cache, error) {
	// Set default values if not provided
	if config.PoolSize == 0 {
		config.PoolSize = 10
	}
	if config.MinIdleConns == 0 {
		config.MinIdleConns = 2
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.DialTimeout == 0 {
		config.DialTimeout = 5 * time.Second
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 3 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 3 * time.Second
	}

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Redis cache connection established with pool size %d", config.PoolSize)

	return &Cache{
		client: rdb,
		config: config,
	}, nil
}

// Set stores a value in cache with expiration
func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set cache key %s: %w", key, err)
	}

	return nil
}

// Get retrieves a value from cache
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get cache key %s: %w", key, err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	return nil
}

// Delete removes a key from cache
func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete cache keys: %w", err)
	}

	return nil
}

// Exists checks if a key exists in cache
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache key existence: %w", err)
	}

	return count > 0, nil
}

// SetNX sets a key only if it doesn't exist (atomic operation)
func (c *Cache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	result, err := c.client.SetNX(ctx, key, data, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set cache key %s: %w", key, err)
	}

	return result, nil
}

// Increment atomically increments a counter
func (c *Cache) Increment(ctx context.Context, key string) (int64, error) {
	result, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment cache key %s: %w", key, err)
	}

	return result, nil
}

// IncrementWithExpiry atomically increments a counter with expiration
func (c *Cache) IncrementWithExpiry(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	pipe := c.client.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("failed to increment cache key %s with expiry: %w", key, err)
	}

	return incrCmd.Val(), nil
}

// GetOrSet retrieves a value from cache, or sets it if not found
func (c *Cache) GetOrSet(ctx context.Context, key string, dest interface{}, setter func() (interface{}, error), expiration time.Duration) error {
	// Try to get from cache first
	err := c.Get(ctx, key, dest)
	if err == nil {
		return nil // Found in cache
	}

	if err != ErrCacheMiss {
		return err // Real error occurred
	}

	// Not in cache, call setter function
	value, err := setter()
	if err != nil {
		return fmt.Errorf("setter function failed: %w", err)
	}

	// Store in cache
	if err := c.Set(ctx, key, value, expiration); err != nil {
		log.Printf("Failed to cache value for key %s: %v", key, err)
		// Don't return error here, just log it
	}

	// Marshal the value to dest
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal setter result: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal setter result: %w", err)
	}

	return nil
}

// InvalidatePattern deletes all keys matching a pattern
func (c *Cache) InvalidatePattern(ctx context.Context, pattern string) error {
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}

	if len(keys) == 0 {
		return nil
	}

	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys for pattern %s: %w", pattern, err)
	}

	log.Printf("Invalidated %d cache keys matching pattern: %s", len(keys), pattern)
	return nil
}

// HealthCheck performs a health check on Redis
func (c *Cache) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

// Stats returns Redis connection pool statistics
func (c *Cache) Stats() *redis.PoolStats {
	return c.client.PoolStats()
}

// Close closes the Redis connection
func (c *Cache) Close() error {
	log.Println("Closing Redis cache connection")
	return c.client.Close()
}

// CacheKey generates a cache key with prefix
func CacheKey(prefix string, parts ...string) string {
	key := prefix
	for _, part := range parts {
		key += ":" + part
	}
	return key
}

// Common cache key prefixes
const (
	DestinationPrefix = "destination"
	UserPrefix        = "user"
	BookingPrefix     = "booking"
	ReviewPrefix      = "review"
	SessionPrefix     = "session"
	RateLimitPrefix   = "rate_limit"
)

// Common cache durations
const (
	ShortTTL  = 5 * time.Minute
	MediumTTL = 30 * time.Minute
	LongTTL   = 2 * time.Hour
	DayTTL    = 24 * time.Hour
)

// ErrCacheMiss is returned when a cache key is not found
var ErrCacheMiss = fmt.Errorf("cache miss")

// CacheManager provides high-level caching operations
type CacheManager struct {
	cache *Cache
}

// NewCacheManager creates a new cache manager
func NewCacheManager(cache *Cache) *CacheManager {
	return &CacheManager{cache: cache}
}

// CacheDestination caches a destination
func (cm *CacheManager) CacheDestination(ctx context.Context, destinationID string, destination interface{}) error {
	key := CacheKey(DestinationPrefix, destinationID)
	return cm.cache.Set(ctx, key, destination, LongTTL)
}

// GetDestination retrieves a destination from cache
func (cm *CacheManager) GetDestination(ctx context.Context, destinationID string, dest interface{}) error {
	key := CacheKey(DestinationPrefix, destinationID)
	return cm.cache.Get(ctx, key, dest)
}

// InvalidateDestination removes a destination from cache
func (cm *CacheManager) InvalidateDestination(ctx context.Context, destinationID string) error {
	key := CacheKey(DestinationPrefix, destinationID)
	return cm.cache.Delete(ctx, key)
}

// CacheUser caches a user
func (cm *CacheManager) CacheUser(ctx context.Context, userID string, user interface{}) error {
	key := CacheKey(UserPrefix, userID)
	return cm.cache.Set(ctx, key, user, MediumTTL)
}

// GetUser retrieves a user from cache
func (cm *CacheManager) GetUser(ctx context.Context, userID string, dest interface{}) error {
	key := CacheKey(UserPrefix, userID)
	return cm.cache.Get(ctx, key, dest)
}

// InvalidateUser removes a user from cache
func (cm *CacheManager) InvalidateUser(ctx context.Context, userID string) error {
	key := CacheKey(UserPrefix, userID)
	return cm.cache.Delete(ctx, key)
}

// RateLimitCheck checks and updates rate limit counter
func (cm *CacheManager) RateLimitCheck(ctx context.Context, identifier string, limit int64, window time.Duration) (bool, error) {
	key := CacheKey(RateLimitPrefix, identifier)
	
	count, err := cm.cache.IncrementWithExpiry(ctx, key, window)
	if err != nil {
		return false, err
	}

	return count <= limit, nil
}

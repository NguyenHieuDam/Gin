package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// GetRedisConfig returns Redis configuration from environment variables
func GetRedisConfig() *RedisConfig {
	return &RedisConfig{
		Addr:     getRedisEnv("REDIS_HOST", "localhost") + ":" + getRedisEnv("REDIS_PORT", "6379"),
		Password: getRedisEnv("REDIS_PASSWORD", ""),
		DB:       getRedisEnvInt("REDIS_DB", 0),
	}
}

// ConnectRedis establishes connection to Redis
func ConnectRedis() (*redis.Client, error) {
	config := GetRedisConfig()
	
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
		
		// Connection pool settings
		PoolSize:     10,
		MinIdleConns: 5,
		
		// Connection timeout settings
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		
		// Retry settings
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("✅ Connected to Redis successfully")
	return rdb, nil
}

// CloseRedis closes the Redis connection
func CloseRedis(rdb *redis.Client) error {
	return rdb.Close()
}

// HealthCheck checks Redis connectivity
func HealthCheckRedis(rdb *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	_, err := rdb.Ping(ctx).Result()
	return err
}

// Redis utilities for common operations

// SetWithExpiry sets a key-value pair with expiration
func SetWithExpiry(rdb *redis.Client, key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()
	return rdb.Set(ctx, key, value, expiration).Err()
}

// GetValue gets a value by key
func GetValue(rdb *redis.Client, key string) (string, error) {
	ctx := context.Background()
	return rdb.Get(ctx, key).Result()
}

// DeleteKey deletes a key
func DeleteKey(rdb *redis.Client, key string) error {
	ctx := context.Background()
	return rdb.Del(ctx, key).Err()
}

// IncrementCounter increments a counter with expiration
func IncrementCounter(rdb *redis.Client, key string, expiration time.Duration) (int64, error) {
	ctx := context.Background()
	pipe := rdb.Pipeline()
	
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	
	return incr.Val(), nil
}

// AddToSet adds a member to a set
func AddToSet(rdb *redis.Client, key string, members ...interface{}) error {
	ctx := context.Background()
	return rdb.SAdd(ctx, key, members...).Err()
}

// RemoveFromSet removes a member from a set
func RemoveFromSet(rdb *redis.Client, key string, members ...interface{}) error {
	ctx := context.Background()
	return rdb.SRem(ctx, key, members...).Err()
}

// GetSetMembers gets all members of a set
func GetSetMembers(rdb *redis.Client, key string) ([]string, error) {
	ctx := context.Background()
	return rdb.SMembers(ctx, key).Result()
}

// SetHash sets hash fields
func SetHash(rdb *redis.Client, key string, values map[string]interface{}) error {
	ctx := context.Background()
	return rdb.HMSet(ctx, key, values).Err()
}

// GetHash gets all hash fields
func GetHash(rdb *redis.Client, key string) (map[string]string, error) {
	ctx := context.Background()
	return rdb.HGetAll(ctx, key).Result()
}

// getRedisEnv gets environment variable with fallback for Redis
func getRedisEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getRedisEnvInt gets environment variable as int with fallback for Redis
func getRedisEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Sscanf(value, "%d", &fallback); err == nil && intValue == 1 {
			return fallback
		}
	}
	return fallback
}

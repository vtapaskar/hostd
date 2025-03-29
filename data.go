package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

// RedisClient wraps Redis operations
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis client
func NewRedisClient(config *RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisClient{
		client: client,
	}, nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// UpdateProcessStatus updates the status of a process in Redis
func (r *RedisClient) UpdateProcessStatus(ctx context.Context, processName string, status string) error {
	key := fmt.Sprintf("process:%s:status", processName)
	return r.client.Set(ctx, key, status, 0).Err()
}

// GetProcessStatus gets the status of a process from Redis
func (r *RedisClient) GetProcessStatus(ctx context.Context, processName string) (string, error) {
	key := fmt.Sprintf("process:%s:status", processName)
	return r.client.Get(ctx, key).Result()
}

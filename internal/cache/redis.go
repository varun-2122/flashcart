package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/varun-2122/flashcart/internal/config"
	"github.com/varun-2122/flashcart/internal/logger"
)

// RedisClient manages Redis cache connections.
type RedisClient struct {
	Client *redis.Client
	mu     sync.RWMutex
}

// NewRedisClient initializes and pings Redis cache.
func NewRedisClient(ctx context.Context, cfg *config.CacheConfig) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: 2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	logger.Info(ctx, "successfully connected to redis cache",
		"addr", cfg.Addr(),
		"db", cfg.DB,
		"pool_size", cfg.PoolSize,
	)

	return &RedisClient{Client: rdb}, nil
}

// Ping checks if Redis is responsive.
func (r *RedisClient) Ping(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.Client == nil {
		return fmt.Errorf("redis client is not initialized")
	}
	return r.Client.Ping(ctx).Err()
}

// Close gracefully closes connection to Redis.
func (r *RedisClient) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Client != nil {
		err := r.Client.Close()
		r.Client = nil
		return err
	}
	return nil
}

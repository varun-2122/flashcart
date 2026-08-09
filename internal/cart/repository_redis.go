package cart

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/cache"
	"github.com/varun-2122/flashcart/internal/domain"
)

type RedisCartRepository struct {
	redis *cache.RedisClient
	ttl   time.Duration
}

func NewRedisCartRepository(redis *cache.RedisClient) domain.CartRepository {
	return &RedisCartRepository{
		redis: redis,
		ttl:   7 * 24 * time.Hour,
	}
}

func (r *RedisCartRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Cart, error) {
	if r.redis == nil || r.redis.Client == nil {
		return &domain.Cart{UserID: userID, Items: []domain.CartItem{}, UpdatedAt: time.Now()}, nil
	}

	key := fmt.Sprintf("cart:%s", userID.String())
	val, err := r.redis.Client.Get(ctx, key).Result()
	if err != nil || val == "" {
		return &domain.Cart{UserID: userID, Items: []domain.CartItem{}, UpdatedAt: time.Now()}, nil
	}

	var cart domain.Cart
	if err := json.Unmarshal([]byte(val), &cart); err != nil {
		return &domain.Cart{UserID: userID, Items: []domain.CartItem{}, UpdatedAt: time.Now()}, nil
	}

	return &cart, nil
}

func (r *RedisCartRepository) Save(ctx context.Context, cart *domain.Cart) error {
	if r.redis == nil || r.redis.Client == nil {
		return nil
	}

	cart.UpdatedAt = time.Now()
	bytes, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("failed to marshal cart: %w", err)
	}

	key := fmt.Sprintf("cart:%s", cart.UserID.String())
	return r.redis.Client.Set(ctx, key, bytes, r.ttl).Err()
}

func (r *RedisCartRepository) Clear(ctx context.Context, userID uuid.UUID) error {
	if r.redis == nil || r.redis.Client == nil {
		return nil
	}
	key := fmt.Sprintf("cart:%s", userID.String())
	return r.redis.Client.Del(ctx, key).Err()
}

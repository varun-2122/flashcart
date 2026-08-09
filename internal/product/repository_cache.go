package product

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/cache"
	"github.com/varun-2122/flashcart/internal/domain"
	"github.com/varun-2122/flashcart/internal/logger"
)

type CachedProductRepository struct {
	repo  domain.ProductRepository
	redis *cache.RedisClient
	ttl   time.Duration
}

func NewCachedProductRepository(repo domain.ProductRepository, redis *cache.RedisClient) domain.ProductRepository {
	return &CachedProductRepository{
		repo:  repo,
		redis: redis,
		ttl:   1 * time.Hour,
	}
}

func (r *CachedProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	if r.redis != nil && r.redis.Client != nil {
		cacheKey := fmt.Sprintf("product:%s", id.String())
		val, err := r.redis.Client.Get(ctx, cacheKey).Result()
		if err == nil && val != "" {
			var p domain.Product
			if err := json.Unmarshal([]byte(val), &p); err == nil {
				logger.Debug(ctx, "product cache hit", "product_id", id.String())
				return &p, nil
			}
		}
	}

	p, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if r.redis != nil && r.redis.Client != nil {
		cacheKey := fmt.Sprintf("product:%s", id.String())
		bytes, err := json.Marshal(p)
		if err == nil {
			_ = r.redis.Client.Set(ctx, cacheKey, bytes, r.ttl).Err()
		}
	}

	return p, nil
}

func (r *CachedProductRepository) Create(ctx context.Context, p *domain.Product) error {
	return r.repo.Create(ctx, p)
}

func (r *CachedProductRepository) List(ctx context.Context, filter domain.ProductFilter) ([]*domain.Product, int, error) {
	return r.repo.List(ctx, filter)
}

func (r *CachedProductRepository) Update(ctx context.Context, p *domain.Product) error {
	err := r.repo.Update(ctx, p)
	if err == nil && r.redis != nil && r.redis.Client != nil {
		cacheKey := fmt.Sprintf("product:%s", p.ID.String())
		_ = r.redis.Client.Del(ctx, cacheKey).Err()
	}
	return err
}

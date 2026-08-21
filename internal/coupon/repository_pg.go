package coupon

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/varun-2122/flashcart/internal/database"
	"github.com/varun-2122/flashcart/internal/domain"
)

// PostgresCouponRepository implements domain.CouponRepository.
type PostgresCouponRepository struct {
	db *database.PostgresDB
}

// NewPostgresCouponRepository constructs the coupon repository.
func NewPostgresCouponRepository(db *database.PostgresDB) domain.CouponRepository {
	return &PostgresCouponRepository{db: db}
}

func (r *PostgresCouponRepository) GetByCode(ctx context.Context, code string) (*domain.Coupon, error) {
	query := `
		SELECT id, code, discount_percent, max_uses, used_count, expires_at, is_active, created_at
		FROM coupons WHERE code = $1
	`
	c := &domain.Coupon{}
	err := r.db.Pool.QueryRow(ctx, query, code).Scan(
		&c.ID, &c.Code, &c.DiscountPercent, &c.MaxUses, &c.UsedCount, &c.ExpiresAt, &c.IsActive, &c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCouponNotFound
		}
		return nil, fmt.Errorf("failed to get coupon by code: %w", err)
	}
	return c, nil
}

func (r *PostgresCouponRepository) GetUserCouponUse(ctx context.Context, couponID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM coupon_uses WHERE coupon_id = $1 AND user_id = $2)`
	var used bool
	if err := r.db.Pool.QueryRow(ctx, query, couponID, userID).Scan(&used); err != nil {
		return false, fmt.Errorf("failed to check coupon use: %w", err)
	}
	return used, nil
}

func (r *PostgresCouponRepository) RecordUse(ctx context.Context, couponID, userID, orderID uuid.UUID) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin coupon use transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Insert use record
	_, err = tx.Exec(ctx,
		`INSERT INTO coupon_uses (id, coupon_id, user_id, order_id) VALUES ($1, $2, $3, $4)`,
		uuid.New(), couponID, userID, orderID,
	)
	if err != nil {
		return fmt.Errorf("failed to record coupon use: %w", err)
	}

	// Increment used_count atomically
	_, err = tx.Exec(ctx,
		`UPDATE coupons SET used_count = used_count + 1 WHERE id = $1`,
		couponID,
	)
	if err != nil {
		return fmt.Errorf("failed to increment coupon used_count: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PostgresCouponRepository) Create(ctx context.Context, c *domain.Coupon) error {
	query := `
		INSERT INTO coupons (id, code, discount_percent, max_uses, used_count, expires_at, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		c.ID, c.Code, c.DiscountPercent, c.MaxUses, c.UsedCount, c.ExpiresAt, c.IsActive, c.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create coupon: %w", err)
	}
	return nil
}

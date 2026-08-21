package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCouponNotFound    = errors.New("coupon code not found")
	ErrCouponExpired     = errors.New("coupon has expired")
	ErrCouponExhausted   = errors.New("coupon maximum usage limit reached")
	ErrCouponInactive    = errors.New("coupon is not active")
	ErrCouponAlreadyUsed = errors.New("you have already used this coupon")
)

// Coupon represents a discount code.
type Coupon struct {
	ID              uuid.UUID `json:"id"`
	Code            string    `json:"code"`
	DiscountPercent float64   `json:"discount_percent"`
	MaxUses         int       `json:"max_uses"`
	UsedCount       int       `json:"used_count"`
	ExpiresAt       time.Time `json:"expires_at"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}

// IsValid checks whether this coupon can be applied.
func (c *Coupon) IsValid() error {
	if !c.IsActive {
		return ErrCouponInactive
	}
	if time.Now().After(c.ExpiresAt) {
		return ErrCouponExpired
	}
	if c.MaxUses > 0 && c.UsedCount >= c.MaxUses {
		return ErrCouponExhausted
	}
	return nil
}

// ApplyDiscount returns the discounted amount.
func (c *Coupon) ApplyDiscount(original float64) float64 {
	discount := original * (c.DiscountPercent / 100.0)
	return original - discount
}

// CouponRepository defines data access for coupons.
type CouponRepository interface {
	GetByCode(ctx context.Context, code string) (*Coupon, error)
	GetUserCouponUse(ctx context.Context, couponID, userID uuid.UUID) (bool, error)
	RecordUse(ctx context.Context, couponID, userID, orderID uuid.UUID) error
	Create(ctx context.Context, coupon *Coupon) error
}

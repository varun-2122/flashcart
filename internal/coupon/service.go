package coupon

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/domain"
	"github.com/varun-2122/flashcart/internal/metrics"
)

// ValidateResponse is the result of coupon validation.
type ValidateResponse struct {
	Valid           bool    `json:"valid"`
	Code            string  `json:"code"`
	DiscountPercent float64 `json:"discount_percent"`
	Message         string  `json:"message"`
}

// CouponService handles coupon validation and application.
type CouponService struct {
	repo domain.CouponRepository
}

// NewCouponService creates a CouponService.
func NewCouponService(repo domain.CouponRepository) *CouponService {
	return &CouponService{repo: repo}
}

// Validate checks if a coupon code is valid for the given user.
// It does NOT consume the coupon — call RecordUse after order creation.
func (s *CouponService) Validate(ctx context.Context, code string, userID uuid.UUID) (*ValidateResponse, error) {
	coupon, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		metrics.CouponsValidated.WithLabelValues("not_found").Inc()
		return &ValidateResponse{Valid: false, Message: "Coupon code not found"}, nil
	}

	if err := coupon.IsValid(); err != nil {
		label := "invalid"
		switch err {
		case domain.ErrCouponExpired:
			label = "expired"
		case domain.ErrCouponExhausted:
			label = "exhausted"
		case domain.ErrCouponInactive:
			label = "inactive"
		}
		metrics.CouponsValidated.WithLabelValues(label).Inc()
		return &ValidateResponse{Valid: false, Message: err.Error()}, nil
	}

	// Check per-user usage
	used, err := s.repo.GetUserCouponUse(ctx, coupon.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check coupon usage history: %w", err)
	}
	if used {
		metrics.CouponsValidated.WithLabelValues("already_used").Inc()
		return &ValidateResponse{Valid: false, Message: domain.ErrCouponAlreadyUsed.Error()}, nil
	}

	metrics.CouponsValidated.WithLabelValues("valid").Inc()
	return &ValidateResponse{
		Valid:           true,
		Code:            coupon.Code,
		DiscountPercent: coupon.DiscountPercent,
		Message:         fmt.Sprintf("%.0f%% discount applied", coupon.DiscountPercent),
	}, nil
}

// RecordUse persists a coupon use after a successful order.
func (s *CouponService) RecordUse(ctx context.Context, code string, userID, orderID uuid.UUID) error {
	coupon, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return err
	}
	return s.repo.RecordUse(ctx, coupon.ID, userID, orderID)
}

// GetByCode returns a coupon entity by its code.
func (s *CouponService) GetByCode(ctx context.Context, code string) (*domain.Coupon, error) {
	return s.repo.GetByCode(ctx, code)
}

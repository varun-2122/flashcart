package payment

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/domain"
	"github.com/varun-2122/flashcart/internal/logger"
	"github.com/varun-2122/flashcart/internal/metrics"
)

// ChargeRequest is the input to the payment service.
type ChargeRequest struct {
	OrderID        uuid.UUID
	UserID         uuid.UUID
	Amount         float64
	IdempotencyKey string
	Provider       string // e.g. "stripe", "razorpay", "simulated"
}

// PaymentService handles payment lifecycle with idempotency protection.
type PaymentService struct {
	repo domain.PaymentRepository
}

// NewPaymentService creates a PaymentService backed by the given repository.
func NewPaymentService(repo domain.PaymentRepository) *PaymentService {
	return &PaymentService{repo: repo}
}

// Charge initiates a payment for an order. Idempotent — safe to retry.
//
// Flow:
//  1. Check if payment already exists for the idempotency key → return cached result.
//  2. Persist a PENDING payment record.
//  3. Simulate provider charge (success / failure with 10% failure rate in sim).
//  4. Update status to CAPTURED or FAILED.
func (s *PaymentService) Charge(ctx context.Context, req ChargeRequest) (*domain.Payment, error) {
	if req.IdempotencyKey == "" {
		req.IdempotencyKey = fmt.Sprintf("%s-%s", req.OrderID.String(), req.UserID.String())
	}

	if req.Provider == "" {
		req.Provider = "simulated"
	}

	// 1. Idempotency check — return existing result if already processed.
	existing, err := s.repo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if err == nil && existing != nil {
		logger.Info(ctx, "payment already processed, returning cached result",
			"payment_id", existing.ID.String(),
			"idempotency_key", req.IdempotencyKey,
		)
		return existing, nil
	}

	// 2. Create PENDING payment record.
	now := time.Now()
	payment := &domain.Payment{
		ID:             uuid.New(),
		OrderID:        req.OrderID,
		UserID:         req.UserID,
		Amount:         req.Amount,
		Status:         domain.PaymentStatusPending,
		Provider:       req.Provider,
		IdempotencyKey: req.IdempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	logger.Info(ctx, "payment initiated",
		"payment_id", payment.ID.String(),
		"order_id", req.OrderID.String(),
		"amount", req.Amount,
		"provider", req.Provider,
	)

	// 3. Simulate provider charge (replace this with real Stripe/Razorpay call).
	txID, providerErr := simulateProviderCharge(req.Provider, req.Amount)

	// 4. Update payment status based on provider response.
	if providerErr != nil {
		_ = s.repo.UpdateStatus(ctx, payment.ID, domain.PaymentStatusFailed, "", providerErr.Error())
		payment.Status = domain.PaymentStatusFailed
		payment.FailureReason = providerErr.Error()

		metrics.PaymentsTotal.WithLabelValues("failed").Inc()
		logger.Error(ctx, "payment failed",
			"payment_id", payment.ID.String(),
			"reason", providerErr.Error(),
		)
		return payment, domain.ErrPaymentFailed
	}

	_ = s.repo.UpdateStatus(ctx, payment.ID, domain.PaymentStatusCaptured, txID, "")
	payment.Status = domain.PaymentStatusCaptured
	payment.TransactionID = txID

	metrics.PaymentsTotal.WithLabelValues("captured").Inc()
	metrics.PaymentAmount.Observe(req.Amount)

	logger.Info(ctx, "payment captured",
		"payment_id", payment.ID.String(),
		"transaction_id", txID,
		"amount", req.Amount,
	)

	return payment, nil
}

// GetByOrderID returns the latest payment for an order.
func (s *PaymentService) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.Payment, error) {
	return s.repo.GetByOrderID(ctx, orderID)
}

// GetByID returns a payment by its ID.
func (s *PaymentService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	return s.repo.GetByID(ctx, id)
}

// simulateProviderCharge pretends to call a payment gateway.
// Replace with a real provider SDK (stripe-go, razorpay-go, etc.).
// Returns (transactionID, error). Simulates ~10% failure rate.
func simulateProviderCharge(provider string, amount float64) (string, error) {
	// Simulate network latency
	time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

	// 10% failure rate simulation for testing error handling
	if rand.Float32() < 0.10 {
		return "", fmt.Errorf("provider %s: card declined (simulated failure)", provider)
	}

	txID := fmt.Sprintf("%s_txn_%s", provider, uuid.New().String()[:8])
	return txID, nil
}

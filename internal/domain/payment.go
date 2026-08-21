package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// PaymentStatus represents the lifecycle state of a payment.
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusCaptured  PaymentStatus = "CAPTURED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusRefunded  PaymentStatus = "REFUNDED"
)

var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrPaymentAlreadyExists = errors.New("payment already processed for this idempotency key")
	ErrPaymentFailed        = errors.New("payment processing failed")
)

// Payment represents a payment attempt for an order.
type Payment struct {
	ID             uuid.UUID     `json:"id"`
	OrderID        uuid.UUID     `json:"order_id"`
	UserID         uuid.UUID     `json:"user_id"`
	Amount         float64       `json:"amount"`
	Status         PaymentStatus `json:"status"`
	Provider       string        `json:"provider"`
	IdempotencyKey string        `json:"-"`
	TransactionID  string        `json:"transaction_id,omitempty"`
	FailureReason  string        `json:"failure_reason,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// PaymentRepository defines data access contracts for payments.
type PaymentRepository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*Payment, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*Payment, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status PaymentStatus, txID, reason string) error
}

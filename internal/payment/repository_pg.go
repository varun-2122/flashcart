package payment

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/varun-2122/flashcart/internal/database"
	"github.com/varun-2122/flashcart/internal/domain"
)

// PostgresPaymentRepository implements domain.PaymentRepository using pgx.
type PostgresPaymentRepository struct {
	db *database.PostgresDB
}

// NewPostgresPaymentRepository constructs a new payment repository.
func NewPostgresPaymentRepository(db *database.PostgresDB) domain.PaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

func (r *PostgresPaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, user_id, amount, status, provider, idempotency_key, transaction_id, failure_reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		p.ID, p.OrderID, p.UserID, p.Amount, p.Status, p.Provider,
		p.IdempotencyKey, p.TransactionID, p.FailureReason, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}
	return nil
}

func (r *PostgresPaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, user_id, amount, status, provider, idempotency_key, transaction_id, failure_reason, created_at, updated_at
		FROM payments WHERE id = $1
	`
	p := &domain.Payment{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Status, &p.Provider,
		&p.IdempotencyKey, &p.TransactionID, &p.FailureReason, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to get payment by id: %w", err)
	}
	return p, nil
}

func (r *PostgresPaymentRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, user_id, amount, status, provider, idempotency_key, transaction_id, failure_reason, created_at, updated_at
		FROM payments WHERE order_id = $1 ORDER BY created_at DESC LIMIT 1
	`
	p := &domain.Payment{}
	err := r.db.Pool.QueryRow(ctx, query, orderID).Scan(
		&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Status, &p.Provider,
		&p.IdempotencyKey, &p.TransactionID, &p.FailureReason, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to get payment by order id: %w", err)
	}
	return p, nil
}

func (r *PostgresPaymentRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, user_id, amount, status, provider, idempotency_key, transaction_id, failure_reason, created_at, updated_at
		FROM payments WHERE idempotency_key = $1
	`
	p := &domain.Payment{}
	err := r.db.Pool.QueryRow(ctx, query, key).Scan(
		&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Status, &p.Provider,
		&p.IdempotencyKey, &p.TransactionID, &p.FailureReason, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to get payment by idempotency key: %w", err)
	}
	return p, nil
}

func (r *PostgresPaymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus, txID, reason string) error {
	query := `
		UPDATE payments
		SET status = $1, transaction_id = $2, failure_reason = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.db.Pool.Exec(ctx, query, status, txID, reason, id)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}
	return nil
}

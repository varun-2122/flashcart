package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/varun-2122/flashcart/internal/database"
	"github.com/varun-2122/flashcart/internal/domain"
)

type PostgresOrderRepository struct {
	db *database.PostgresDB
}

func NewPostgresOrderRepository(db *database.PostgresDB) domain.OrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, o *domain.Order) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin order transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	orderQuery := `
		INSERT INTO orders (id, user_id, total_amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if _, err := tx.Exec(ctx, orderQuery, o.ID, o.UserID, o.TotalAmount, o.Status, o.CreatedAt, o.UpdatedAt); err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	itemQuery := `
		INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	for _, item := range o.Items {
		if _, err := tx.Exec(ctx, itemQuery, item.ID, o.ID, item.ProductID, item.Quantity, item.UnitPrice, item.CreatedAt); err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit order transaction: %w", err)
	}
	return nil
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	query := `
		SELECT id, user_id, total_amount, status, created_at, updated_at
		FROM orders WHERE id = $1
	`
	o := &domain.Order{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&o.ID, &o.UserID, &o.TotalAmount, &o.Status, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to query order: %w", err)
	}

	itemsQuery := `
		SELECT id, order_id, product_id, quantity, unit_price, created_at
		FROM order_items WHERE order_id = $1
	`
	rows, err := r.db.Pool.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		item := domain.OrderItem{}
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.UnitPrice, &item.CreatedAt); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, item)
	}

	return o, nil
}

func (r *PostgresOrderRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Order, error) {
	query := `
		SELECT id, user_id, total_amount, status, created_at, updated_at
		FROM orders WHERE user_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user orders: %w", err)
	}
	defer rows.Close()

	orders := make([]*domain.Order, 0)
	for rows.Next() {
		o := &domain.Order{}
		if err := rows.Scan(&o.ID, &o.UserID, &o.TotalAmount, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, nil
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	return nil
}

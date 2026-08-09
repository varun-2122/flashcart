package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusPaid      OrderStatus = "PAID"
	OrderStatusShipped   OrderStatus = "SHIPPED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrEmptyCart     = errors.New("cannot create order from empty cart")
)

// OrderItem represents a line item in an order.
type OrderItem struct {
	ID        uuid.UUID `json:"id"`
	OrderID   uuid.UUID `json:"order_id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
	CreatedAt time.Time `json:"created_at"`
}

// Order represents an order domain entity.
type Order struct {
	ID          uuid.UUID   `json:"id"`
	UserID      uuid.UUID   `json:"user_id"`
	TotalAmount float64     `json:"total_amount"`
	Status      OrderStatus `json:"status"`
	Items       []OrderItem `json:"items,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// OrderRepository defines data operations for orders.
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status OrderStatus) error
}

package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// CartItem represents an item in shopping cart.
type CartItem struct {
	ProductID uuid.UUID `json:"product_id"`
	Name      string    `json:"name"`
	UnitPrice float64   `json:"unit_price"`
	Quantity  int       `json:"quantity"`
}

// Cart represents a user's shopping cart.
type Cart struct {
	UserID    uuid.UUID  `json:"user_id"`
	Items     []CartItem `json:"items"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Total computes total amount for cart.
func (c *Cart) Total() float64 {
	var total float64
	for _, item := range c.Items {
		total += item.UnitPrice * float64(item.Quantity)
	}
	return total
}

// CartRepository defines cart storage contracts.
type CartRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Cart, error)
	Save(ctx context.Context, cart *Cart) error
	Clear(ctx context.Context, userID uuid.UUID) error
}

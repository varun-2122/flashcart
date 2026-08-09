package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidSKU      = errors.New("product SKU must be unique and non-empty")
)

// Category represents a product category.
type Category struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

// Product represents a product catalog entity.
type Product struct {
	ID          uuid.UUID  `json:"id"`
	SKU         string     `json:"sku"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Price       float64    `json:"price"`
	CategoryID  *uuid.UUID `json:"category_id,omitempty"`
	Brand       string     `json:"brand"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ProductFilter parameters for searching and pagination.
type ProductFilter struct {
	CategoryID *uuid.UUID
	Brand      string
	MinPrice   *float64
	MaxPrice   *float64
	Search     string
	Limit      int
	Offset     int
}

// ProductRepository defines storage contracts for catalog.
type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	List(ctx context.Context, filter ProductFilter) ([]*Product, int, error)
	Update(ctx context.Context, product *Product) error
}

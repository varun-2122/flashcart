package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInventoryNotFound        = errors.New("inventory record not found for product")
	ErrInsufficientStock        = errors.New("insufficient inventory stock available")
	ErrOptimisticLockConflict = errors.New("optimistic locking conflict: stock updated by another transaction")
)

// Inventory represents stock quantity and optimistic locking version.
type Inventory struct {
	ProductID        uuid.UUID `json:"product_id"`
	Quantity         int       `json:"quantity"`
	ReservedQuantity int       `json:"reserved_quantity"`
	Version          int64     `json:"version"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// AvailableQuantity returns unreserved stock.
func (i *Inventory) AvailableQuantity() int {
	return i.Quantity - i.ReservedQuantity
}

// InventoryRepository defines stock management and optimistic locking contracts.
type InventoryRepository interface {
	GetByProductID(ctx context.Context, productID uuid.UUID) (*Inventory, error)
	SetStock(ctx context.Context, productID uuid.UUID, quantity int) error
	ReserveStockWithOptimisticLock(ctx context.Context, productID uuid.UUID, quantity int, currentVersion int64) error
}

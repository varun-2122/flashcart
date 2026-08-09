package inventory

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/varun-2122/flashcart/internal/database"
	"github.com/varun-2122/flashcart/internal/domain"
)

type PostgresInventoryRepository struct {
	db *database.PostgresDB
}

func NewPostgresInventoryRepository(db *database.PostgresDB) domain.InventoryRepository {
	return &PostgresInventoryRepository{db: db}
}

func (r *PostgresInventoryRepository) GetByProductID(ctx context.Context, productID uuid.UUID) (*domain.Inventory, error) {
	query := `
		SELECT product_id, quantity, reserved_quantity, version, updated_at
		FROM inventory WHERE product_id = $1
	`
	inv := &domain.Inventory{}
	err := r.db.Pool.QueryRow(ctx, query, productID).Scan(
		&inv.ProductID, &inv.Quantity, &inv.ReservedQuantity, &inv.Version, &inv.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInventoryNotFound
		}
		return nil, fmt.Errorf("failed to query inventory: %w", err)
	}
	return inv, nil
}

func (r *PostgresInventoryRepository) SetStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	query := `
		INSERT INTO inventory (product_id, quantity, reserved_quantity, version, updated_at)
		VALUES ($1, $2, 0, 1, NOW())
		ON CONFLICT (product_id) DO UPDATE
		SET quantity = EXCLUDED.quantity, version = inventory.version + 1, updated_at = NOW()
	`
	_, err := r.db.Pool.Exec(ctx, query, productID, quantity)
	if err != nil {
		return fmt.Errorf("failed to set inventory stock: %w", err)
	}
	return nil
}

// ReserveStockWithOptimisticLock executes atomic stock deduction with version check.
func (r *PostgresInventoryRepository) ReserveStockWithOptimisticLock(ctx context.Context, productID uuid.UUID, quantity int, currentVersion int64) error {
	query := `
		UPDATE inventory
		SET quantity = quantity - $1, version = version + 1, updated_at = NOW()
		WHERE product_id = $2 AND version = $3 AND (quantity - reserved_quantity) >= $1
	`
	cmd, err := r.db.Pool.Exec(ctx, query, quantity, productID, currentVersion)
	if err != nil {
		return fmt.Errorf("failed to execute optimistic lock stock reservation: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		// Check whether record exists or version mismatched
		existing, err := r.GetByProductID(ctx, productID)
		if err != nil {
			return err
		}
		if existing.AvailableQuantity() < quantity {
			return domain.ErrInsufficientStock
		}
		if existing.Version != currentVersion {
			return domain.ErrOptimisticLockConflict
		}
		return domain.ErrOptimisticLockConflict
	}

	return nil
}

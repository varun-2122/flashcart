package product

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/varun-2122/flashcart/internal/database"
	"github.com/varun-2122/flashcart/internal/domain"
)

type PostgresProductRepository struct {
	db *database.PostgresDB
}

func NewPostgresProductRepository(db *database.PostgresDB) domain.ProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) Create(ctx context.Context, p *domain.Product) error {
	query := `
		INSERT INTO products (id, sku, name, description, price, category_id, brand, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.Pool.Exec(
		ctx, query,
		p.ID, p.SKU, p.Name, p.Description, p.Price, p.CategoryID, p.Brand, p.IsActive, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert product: %w", err)
	}
	return nil
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	query := `
		SELECT id, sku, name, description, price, category_id, brand, is_active, created_at, updated_at
		FROM products WHERE id = $1 AND is_active = TRUE
	`
	p := &domain.Product{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.SKU, &p.Name, &p.Description, &p.Price, &p.CategoryID, &p.Brand, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to query product by id: %w", err)
	}
	return p, nil
}

func (r *PostgresProductRepository) List(ctx context.Context, filter domain.ProductFilter) ([]*domain.Product, int, error) {
	whereClauses := []string{"is_active = TRUE"}
	args := make([]any, 0)
	argIdx := 1

	if filter.CategoryID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("category_id = $%d", argIdx))
		args = append(args, *filter.CategoryID)
		argIdx++
	}

	if filter.Brand != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("brand = $%d", argIdx))
		args = append(args, filter.Brand)
		argIdx++
	}

	if filter.MinPrice != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("price >= $%d", argIdx))
		args = append(args, *filter.MinPrice)
		argIdx++
	}

	if filter.MaxPrice != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("price <= $%d", argIdx))
		args = append(args, *filter.MaxPrice)
		argIdx++
	}

	if filter.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	whereStmt := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products WHERE %s", whereStmt)
	var total int
	if err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT id, sku, name, description, price, category_id, brand, is_active, created_at, updated_at
		FROM products
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereStmt, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	products := make([]*domain.Product, 0)
	for rows.Next() {
		p := &domain.Product{}
		if err := rows.Scan(
			&p.ID, &p.SKU, &p.Name, &p.Description, &p.Price, &p.CategoryID, &p.Brand, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}

	return products, total, nil
}

func (r *PostgresProductRepository) Update(ctx context.Context, p *domain.Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, category_id = $4, brand = $5, is_active = $6, updated_at = $7
		WHERE id = $8
	`
	_, err := r.db.Pool.Exec(ctx, query, p.Name, p.Description, p.Price, p.CategoryID, p.Brand, p.IsActive, p.UpdatedAt, p.ID)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	return nil
}

package database

import (
	"context"
	"fmt"

	"github.com/varun-2122/flashcart/internal/logger"
)

// RunMigrations executes initial SQL migrations to set up schema tables.
func (db *PostgresDB) RunMigrations(ctx context.Context) error {
	if db == nil || db.Pool == nil {
		return fmt.Errorf("postgres pool is not connected")
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'customer',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`,

		`CREATE TABLE IF NOT EXISTS categories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) UNIQUE NOT NULL,
			slug VARCHAR(100) UNIQUE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,

		`CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			sku VARCHAR(100) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price NUMERIC(12, 2) NOT NULL CHECK (price >= 0),
			category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
			brand VARCHAR(100),
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);`,
		`CREATE INDEX IF NOT EXISTS idx_products_brand ON products(brand);`,

		`CREATE TABLE IF NOT EXISTS inventory (
			product_id UUID PRIMARY KEY REFERENCES products(id) ON DELETE CASCADE,
			quantity INT NOT NULL CHECK (quantity >= 0),
			reserved_quantity INT NOT NULL DEFAULT 0 CHECK (reserved_quantity >= 0),
			version BIGINT NOT NULL DEFAULT 1,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,

		`CREATE TABLE IF NOT EXISTS carts (
			user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			items JSONB NOT NULL DEFAULT '[]'::jsonb,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,

		`CREATE TABLE IF NOT EXISTS orders (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			total_amount NUMERIC(12, 2) NOT NULL CHECK (total_amount >= 0),
			status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_orders_user ON orders(user_id);`,

		`CREATE TABLE IF NOT EXISTS order_items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			product_id UUID NOT NULL REFERENCES products(id),
			quantity INT NOT NULL CHECK (quantity > 0),
			unit_price NUMERIC(12, 2) NOT NULL CHECK (unit_price >= 0),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);`,

		`CREATE TABLE IF NOT EXISTS reviews (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
			comment TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(product_id, user_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_reviews_product ON reviews(product_id);`,

		// ── Phase 4: Payments ────────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS payments (
			id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			order_id         UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			user_id          UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
			amount           NUMERIC(12, 2) NOT NULL CHECK (amount >= 0),
			status           VARCHAR(50)  NOT NULL DEFAULT 'PENDING',
			provider         VARCHAR(100) NOT NULL DEFAULT 'simulated',
			idempotency_key  VARCHAR(255) UNIQUE NOT NULL,
			transaction_id   VARCHAR(255),
			failure_reason   TEXT,
			created_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_payments_order    ON payments(order_id);`,
		`CREATE INDEX IF NOT EXISTS idx_payments_user     ON payments(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_payments_idem_key ON payments(idempotency_key);`,

		// ── Phase 4: Coupons ─────────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS coupons (
			id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			code             VARCHAR(100) UNIQUE NOT NULL,
			discount_percent NUMERIC(5, 2) NOT NULL CHECK (discount_percent > 0 AND discount_percent <= 100),
			max_uses         INT NOT NULL DEFAULT 0,
			used_count       INT NOT NULL DEFAULT 0,
			expires_at       TIMESTAMP WITH TIME ZONE NOT NULL,
			is_active        BOOLEAN NOT NULL DEFAULT TRUE,
			created_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons(code);`,

		`CREATE TABLE IF NOT EXISTS coupon_uses (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			coupon_id  UUID NOT NULL REFERENCES coupons(id) ON DELETE CASCADE,
			user_id    UUID NOT NULL REFERENCES users(id)   ON DELETE CASCADE,
			order_id   UUID NOT NULL REFERENCES orders(id)  ON DELETE CASCADE,
			used_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(coupon_id, user_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_coupon_uses_coupon ON coupon_uses(coupon_id);`,
		`CREATE INDEX IF NOT EXISTS idx_coupon_uses_user   ON coupon_uses(user_id);`,
	}

	for _, query := range queries {
		if _, err := db.Pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("migration execution error: %w", err)
		}
	}

	logger.Info(ctx, "database migrations executed successfully")
	return nil
}

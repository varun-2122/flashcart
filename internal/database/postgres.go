package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/varun-2122/flashcart/internal/config"
	"github.com/varun-2122/flashcart/internal/logger"
)

// PostgresDB manages PostgreSQL connection pool.
type PostgresDB struct {
	Pool *pgxpool.Pool
	mu   sync.RWMutex
}

// NewPostgresPool initializes and validates a pgxpool connection pool.
func NewPostgresPool(ctx context.Context, cfg *config.DBConfig) (*PostgresDB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres dsn: %w", err)
	}

	// Apply production connection pool settings
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	// Create pool with context timeout safeguard
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(pingCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres connection pool: %w", err)
	}

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping postgres database: %w", err)
	}

	logger.Info(ctx, "successfully connected to postgresql pool",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.DBName,
		"max_conns", cfg.MaxConns,
		"min_conns", cfg.MinConns,
	)

	return &PostgresDB{Pool: pool}, nil
}

// Ping checks whether PostgreSQL connection pool is alive.
func (db *PostgresDB) Ping(ctx context.Context) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.Pool == nil {
		return fmt.Errorf("postgres pool is not initialized")
	}
	return db.Pool.Ping(ctx)
}

// Close gracefully closes connection pool.
func (db *PostgresDB) Close() {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.Pool != nil {
		db.Pool.Close()
		db.Pool = nil
	}
}

// Stats returns current pool statistics.
func (db *PostgresDB) Stats() *pgxpool.Stat {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.Pool == nil {
		return nil
	}
	return db.Pool.Stat()
}

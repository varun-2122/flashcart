package main

import (
	"context"
	"os"

	"github.com/varun-2122/flashcart/internal/cache"
	"github.com/varun-2122/flashcart/internal/config"
	"github.com/varun-2122/flashcart/internal/database"
	"github.com/varun-2122/flashcart/internal/logger"
	"github.com/varun-2122/flashcart/internal/server"
)

func main() {
	// 1. Load Application Configuration
	cfg, err := config.Load()
	if err != nil {
		println("Fatal: failed to load configuration:", err.Error())
		os.Exit(1)
	}

	// 2. Initialize Structured JSON Logger
	logger.Init(cfg.Logger.Level, cfg.Logger.Format)
	ctx := context.Background()

	logger.Info(ctx, "initializing FlashCart backend engine", "environment", cfg.App.Env)

	// 3. Connect to PostgreSQL Pool
	var db *database.PostgresDB
	db, err = database.NewPostgresPool(ctx, &cfg.Database)
	if err != nil {
		logger.Warn(ctx, "postgresql pool initialization failed (server will operate in degraded state)", "error", err.Error())
	}

	// 4. Connect to Redis Cache
	var redisClient *cache.RedisClient
	redisClient, err = cache.NewRedisClient(ctx, &cfg.Cache)
	if err != nil {
		logger.Warn(ctx, "redis client initialization failed (server will operate in degraded state)", "error", err.Error())
	}

	// 5. Build and Launch Server with Graceful Shutdown Handler
	srv := server.NewServer(cfg, db, redisClient)
	if err := srv.Start(ctx); err != nil {
		logger.Error(ctx, "server execution error", "error", err.Error())
		os.Exit(1)
	}
}

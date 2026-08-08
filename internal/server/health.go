package server

import (
	"context"
	"net/http"
	"time"

	"github.com/varun-2122/flashcart/internal/cache"
	"github.com/varun-2122/flashcart/internal/database"
	"github.com/varun-2122/flashcart/internal/response"
)

// HealthHandler manages system health endpoints.
type HealthHandler struct {
	db    *database.PostgresDB
	redis *cache.RedisClient
}

// NewHealthHandler initializes HealthHandler.
func NewHealthHandler(db *database.PostgresDB, redis *cache.RedisClient) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// Healthz returns basic liveness check.
func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	response.Success(w, map[string]string{
		"status": "UP",
	})
}

// Livez checks liveness.
func (h *HealthHandler) Livez(w http.ResponseWriter, r *http.Request) {
	response.Success(w, map[string]string{
		"status": "ALIVE",
	})
}

// Readyz checks deep component health (Database & Redis).
func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	status := "READY"
	statusCode := http.StatusOK

	dbStatus := "UP"
	if h.db == nil || h.db.Ping(ctx) != nil {
		dbStatus = "DOWN"
		status = "NOT_READY"
		statusCode = http.StatusServiceUnavailable
	}

	redisStatus := "UP"
	if h.redis == nil || h.redis.Ping(ctx) != nil {
		redisStatus = "DOWN"
		status = "NOT_READY"
		statusCode = http.StatusServiceUnavailable
	}

	response.JSON(w, statusCode, map[string]any{
		"status":   status,
		"database": dbStatus,
		"cache":    redisStatus,
		"time":     time.Now().Format(time.RFC3339),
	})
}

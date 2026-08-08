package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/varun-2122/flashcart/internal/cache"
	"github.com/varun-2122/flashcart/internal/config"
	"github.com/varun-2122/flashcart/internal/database"
	"github.com/varun-2122/flashcart/internal/logger"
	"github.com/varun-2122/flashcart/internal/middleware"
)

// Server encapsulates HTTP server state and dependency handles.
type Server struct {
	cfg        *config.Config
	db         *database.PostgresDB
	redis      *cache.RedisClient
	httpServer *http.Server
}

// NewServer builds and configures HTTP server instance.
func NewServer(cfg *config.Config, db *database.PostgresDB, redis *cache.RedisClient) *Server {
	mux := http.NewServeMux()

	healthHandler := NewHealthHandler(db, redis)

	mux.HandleFunc("GET /healthz", healthHandler.Healthz)
	mux.HandleFunc("GET /livez", healthHandler.Livez)
	mux.HandleFunc("GET /readyz", healthHandler.Readyz)

	// API Root Info Endpoint
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"FlashCart API Engine","version":"v1.0.0","status":"operational"}`))
	})

	// Wrap mux with global production middleware chain
	handler := middleware.Chain(
		mux,
		middleware.RequestID,
		middleware.Recovery,
		middleware.CORS,
		middleware.Logger,
		middleware.Timeout(cfg.App.RequestTimeout),
	)

	httpSvr := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      handler,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
		IdleTimeout:  cfg.App.IdleTimeout,
	}

	return &Server{
		cfg:        cfg,
		db:         db,
		redis:      redis,
		httpServer: httpSvr,
	}
}

// Start launches HTTP server and blocks until SIGINT/SIGTERM, then executes graceful shutdown.
func (s *Server) Start(ctx context.Context) error {
	shutdownErr := make(chan error, 1)

	// Run server in background goroutine
	go func() {
		logger.Info(ctx, "starting flashcart HTTP server",
			"port", s.cfg.App.Port,
			"environment", s.cfg.App.Env,
		)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			shutdownErr <- fmt.Errorf("http server failed: %w", err)
		}
	}()

	// Wait for OS shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-shutdownErr:
		return err
	case sig := <-quit:
		logger.Info(ctx, "received shutdown signal", "signal", sig.String())
	}

	// Begin Graceful Shutdown procedure
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.App.ShutdownTimeout)
	defer cancel()

	logger.Info(shutdownCtx, "stopping HTTP server, waiting for active requests to finish...")

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error(shutdownCtx, "forced server shutdown due to timeout", "error", err.Error())
		_ = s.httpServer.Close()
	}

	// Close Database Pool and Redis Client cleanly
	if s.db != nil {
		logger.Info(shutdownCtx, "closing postgresql connection pool...")
		s.db.Close()
	}

	if s.redis != nil {
		logger.Info(shutdownCtx, "closing redis cache client...")
		_ = s.redis.Close()
	}

	logger.Info(shutdownCtx, "flashcart server gracefully stopped")
	return nil
}

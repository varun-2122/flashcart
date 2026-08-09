package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/varun-2122/flashcart/internal/auth"
	"github.com/varun-2122/flashcart/internal/cache"
	"github.com/varun-2122/flashcart/internal/cart"
	"github.com/varun-2122/flashcart/internal/config"
	"github.com/varun-2122/flashcart/internal/database"
	"github.com/varun-2122/flashcart/internal/domain"
	"github.com/varun-2122/flashcart/internal/inventory"
	"github.com/varun-2122/flashcart/internal/logger"
	"github.com/varun-2122/flashcart/internal/middleware"
	"github.com/varun-2122/flashcart/internal/order"
	"github.com/varun-2122/flashcart/internal/product"
	"github.com/varun-2122/flashcart/internal/user"
)

// Server encapsulates HTTP server state and dependency handles.
type Server struct {
	cfg        *config.Config
	db         *database.PostgresDB
	redis      *cache.RedisClient
	httpServer *http.Server
}

// NewServer builds and configures HTTP server instance with Phase 2 domain routes.
func NewServer(cfg *config.Config, db *database.PostgresDB, redis *cache.RedisClient) *Server {
	mux := http.NewServeMux()

	// 1. Core Health Probes
	healthHandler := NewHealthHandler(db, redis)
	mux.HandleFunc("GET /healthz", healthHandler.Healthz)
	mux.HandleFunc("GET /livez", healthHandler.Livez)
	mux.HandleFunc("GET /readyz", healthHandler.Readyz)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"FlashCart API Engine","version":"v2.0.0","status":"operational"}`))
	})

	// 2. Initialize Repositories and Services if database available
	if db != nil && db.Pool != nil {
		// Run DB migrations automatically
		_ = db.RunMigrations(context.Background())

		// Setup Repositories
		userRepo := user.NewPostgresUserRepository(db)
		pgProdRepo := product.NewPostgresProductRepository(db)
		prodRepo := product.NewCachedProductRepository(pgProdRepo, redis)
		invRepo := inventory.NewPostgresInventoryRepository(db)
		cartRepo := cart.NewRedisCartRepository(redis)
		orderRepo := order.NewPostgresOrderRepository(db)

		// Setup Security
		jwtManager := auth.NewJWTManager("", cfg.App.RequestTimeout*100)
		authService := auth.NewAuthService(userRepo, jwtManager)
		authHandler := auth.NewAuthHandler(authService)

		authMiddleware := auth.AuthMiddleware(jwtManager)

		// Setup Domain Services & Handlers
		productService := product.NewProductService(prodRepo, invRepo)
		productHandler := product.NewProductHandler(productService)

		cartService := cart.NewCartService(cartRepo, prodRepo)
		cartHandler := cart.NewCartHandler(cartService)

		orderService := order.NewOrderService(orderRepo, cartRepo, invRepo, prodRepo)
		orderHandler := order.NewOrderHandler(orderService)

		// Auth Routes
		mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
		mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

		// Product Routes
		mux.HandleFunc("GET /api/v1/products", productHandler.ListProducts)
		mux.HandleFunc("GET /api/v1/products/{id}", productHandler.GetProduct)
		mux.Handle("POST /api/v1/products", middleware.Chain(
			http.HandlerFunc(productHandler.CreateProduct),
			authMiddleware,
			auth.RequireRole(domain.RoleAdmin),
		))

		// Cart Routes (Authenticated)
		mux.Handle("GET /api/v1/cart", authMiddleware(http.HandlerFunc(cartHandler.GetCart)))
		mux.Handle("POST /api/v1/cart/items", authMiddleware(http.HandlerFunc(cartHandler.AddItem)))
		mux.Handle("DELETE /api/v1/cart/items/{product_id}", authMiddleware(http.HandlerFunc(cartHandler.RemoveItem)))

		// Order Routes (Authenticated)
		mux.Handle("POST /api/v1/orders", authMiddleware(http.HandlerFunc(orderHandler.CreateOrder)))
		mux.Handle("GET /api/v1/orders", authMiddleware(http.HandlerFunc(orderHandler.ListUserOrders)))
		mux.Handle("GET /api/v1/orders/{id}", authMiddleware(http.HandlerFunc(orderHandler.GetOrder)))
	}

	// Global Production Middleware Chain
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

	go func() {
		logger.Info(ctx, "starting flashcart HTTP server engine",
			"port", s.cfg.App.Port,
			"environment", s.cfg.App.Env,
		)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			shutdownErr <- fmt.Errorf("http server failed: %w", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-shutdownErr:
		return err
	case sig := <-quit:
		logger.Info(ctx, "received shutdown signal", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.App.ShutdownTimeout)
	defer cancel()

	logger.Info(shutdownCtx, "stopping HTTP server, waiting for active requests to finish...")

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error(shutdownCtx, "forced server shutdown due to timeout", "error", err.Error())
		_ = s.httpServer.Close()
	}

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

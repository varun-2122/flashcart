package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	"github.com/varun-2122/flashcart/internal/review"
	"github.com/varun-2122/flashcart/internal/tracing"
	"github.com/varun-2122/flashcart/internal/user"
	"github.com/varun-2122/flashcart/internal/worker"
	"go.opentelemetry.io/otel/sdk/trace"
)

// Server encapsulates HTTP server state and dependency handles.
type Server struct {
	cfg        *config.Config
	db         *database.PostgresDB
	redis      *cache.RedisClient
	httpServer *http.Server
	workerPool *worker.Pool
	tracer     *trace.TracerProvider
}

// NewServer builds and configures HTTP server instance with Phase 3 domain routes.
func NewServer(cfg *config.Config, db *database.PostgresDB, redis *cache.RedisClient) *Server {
	mux := http.NewServeMux()

	// 1. Core Health Probes
	healthHandler := NewHealthHandler(db, redis)
	mux.HandleFunc("GET /healthz", healthHandler.Healthz)
	mux.HandleFunc("GET /livez", healthHandler.Livez)
	mux.HandleFunc("GET /readyz", healthHandler.Readyz)

	// Serve static frontend files from 'web' directory
	fs := http.FileServer(http.Dir("web"))
	mux.Handle("/", fs)

	// 2. Prometheus Metrics Scrape Endpoint
	mux.Handle("GET /metrics", promhttp.Handler())

	// 3. Initialize Worker Pool (5 workers, 100-job buffer)
	pool := worker.NewPool(5, 100)

	// 4. Initialize OpenTelemetry Tracing
	var tracerProvider *trace.TracerProvider
	if tp, err := tracing.InitTracer(context.Background(), "flashcart-api", "localhost:4317"); err == nil {
		tracerProvider = tp
	} else {
		logger.Warn(context.Background(), "failed to initialize opentelemetry tracing", "error", err.Error())
	}

	// 5. Initialize Repositories and Services if database available
	if db != nil && db.Pool != nil {
		// Run DB migrations automatically
		_ = db.RunMigrations(context.Background())

		// Auto-seed database if empty
		seedDatabase(context.Background(), db)

		// Setup Repositories
		userRepo := user.NewPostgresUserRepository(db)
		pgProdRepo := product.NewPostgresProductRepository(db)
		prodRepo := product.NewCachedProductRepository(pgProdRepo, redis)
		invRepo := inventory.NewPostgresInventoryRepository(db)
		cartRepo := cart.NewRedisCartRepository(redis)
		orderRepo := order.NewPostgresOrderRepository(db)
		reviewRepo := review.NewPostgresReviewRepository(db)

		// Setup Security
		jwtManager := auth.NewJWTManager("", cfg.App.RequestTimeout*100)
		authService := auth.NewAuthService(userRepo, jwtManager, cfg.App.GoogleClientID)
		authHandler := auth.NewAuthHandler(authService)

		authMiddleware := auth.AuthMiddleware(jwtManager)

		// Setup Domain Services & Handlers
		productService := product.NewProductService(prodRepo, invRepo)
		productHandler := product.NewProductHandler(productService)

		cartService := cart.NewCartService(cartRepo, prodRepo)
		cartHandler := cart.NewCartHandler(cartService)

		orderService := order.NewOrderService(orderRepo, cartRepo, invRepo, prodRepo, pool)
		orderHandler := order.NewOrderHandler(orderService)

		reviewService := review.NewReviewService(reviewRepo)
		reviewHandler := review.NewReviewHandler(reviewService)

		// Auth Routes
		mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
		mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
		mux.HandleFunc("POST /api/v1/auth/google", authHandler.GoogleLogin)

		// Product Routes
		mux.HandleFunc("GET /api/v1/products", productHandler.ListProducts)
		mux.HandleFunc("GET /api/v1/products/{id}", productHandler.GetProduct)
		mux.Handle("POST /api/v1/products", middleware.Chain(
			http.HandlerFunc(productHandler.CreateProduct),
			authMiddleware,
			auth.RequireRole(domain.RoleAdmin),
		))
		
		// Review Routes
		mux.HandleFunc("GET /api/v1/products/{id}/reviews", reviewHandler.GetReviews)
		mux.Handle("POST /api/v1/products/{id}/reviews", authMiddleware(http.HandlerFunc(reviewHandler.CreateReview)))

		// Cart Routes (Authenticated)
		mux.Handle("GET /api/v1/cart", authMiddleware(http.HandlerFunc(cartHandler.GetCart)))
		mux.Handle("POST /api/v1/cart/items", authMiddleware(http.HandlerFunc(cartHandler.AddItem)))
		mux.Handle("DELETE /api/v1/cart/items/{product_id}", authMiddleware(http.HandlerFunc(cartHandler.RemoveItem)))

		// Order Routes (Authenticated)
		mux.Handle("POST /api/v1/orders", authMiddleware(http.HandlerFunc(orderHandler.CreateOrder)))
		mux.Handle("GET /api/v1/orders", authMiddleware(http.HandlerFunc(orderHandler.ListUserOrders)))
		mux.Handle("GET /api/v1/orders/{id}", authMiddleware(http.HandlerFunc(orderHandler.GetOrder)))
	}

	// Global Production Middleware Chain (Tracing + Metrics added for Phase 4)
	handler := middleware.Chain(
		mux,
		middleware.RequestID,
		middleware.Tracing,
		middleware.Recovery,
		middleware.CORS,
		middleware.Logger,
		middleware.Metrics,
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
		workerPool: pool,
		tracer:     tracerProvider,
	}
}

// Start launches HTTP server and blocks until SIGINT/SIGTERM, then executes graceful shutdown.
func (s *Server) Start(ctx context.Context) error {
	// Start async worker pool goroutines
	s.workerPool.Start(ctx)

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

	// Gracefully drain and stop the worker pool
	s.workerPool.Shutdown(shutdownCtx)

	// Flush tracing buffers
	if s.tracer != nil {
		logger.Info(shutdownCtx, "flushing opentelemetry traces...")
		_ = s.tracer.Shutdown(shutdownCtx)
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

func seedDatabase(ctx context.Context, db *database.PostgresDB) {
	var count int
	err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM products").Scan(&count)
	if err != nil || count > 0 {
		return
	}

	logger.Info(ctx, "database is empty, injecting seed products...")

	products := []domain.Product{
		{
			ID:          uuid.New(),
			SKU:         "TACT-APEX-01",
			Name:        "Apex Daypack",
			Description: "Elite tactical backpack for urban and wilderness environments.",
			Price:       129.99,
			Brand:       "Ignition",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			SKU:         "TACT-CHRON-02",
			Name:        "Chronos Prime Watch",
			Description: "Military-grade tactical smartwatch with GPS and altimeter.",
			Price:       299.50,
			Brand:       "Ignition",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			SKU:         "TACT-GHOST-03",
			Name:        "Ghost Shell Jacket",
			Description: "Waterproof, breathable tactical jacket with hidden compartments.",
			Price:       189.00,
			Brand:       "Ignition",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			SKU:         "TACT-LUMEN-04",
			Name:        "Lumen X Torch",
			Description: "1000 lumen tactical flashlight with strobe and SOS modes.",
			Price:       49.99,
			Brand:       "Ignition",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, p := range products {
		_, err := db.Pool.Exec(ctx, `
			INSERT INTO products (id, sku, name, description, price, brand, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, p.ID, p.SKU, p.Name, p.Description, p.Price, p.Brand, p.IsActive, p.CreatedAt, p.UpdatedAt)

		if err == nil {
			// Insert initial inventory of 100 units
			_, _ = db.Pool.Exec(ctx, `
				INSERT INTO inventory (product_id, available_stock, reserved_stock, version, updated_at)
				VALUES ($1, 100, 0, 1, $2)
			`, p.ID, time.Now())
		}
	}
}

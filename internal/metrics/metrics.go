// Package metrics registers all FlashCart Prometheus business and system metrics.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ── HTTP metrics ─────────────────────────────────────────────────────────────

var (
	// HTTPRequestsTotal counts all HTTP requests by method, path, and status code.
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flashcart_http_requests_total",
			Help: "Total number of HTTP requests processed, partitioned by method, path, and status code.",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration observes request latency distributions per endpoint.
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "flashcart_http_request_duration_seconds",
			Help:    "HTTP request latency distribution in seconds, partitioned by method and path.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

// ── Order business metrics ────────────────────────────────────────────────────

var (
	// OrdersCreated counts all successfully placed orders.
	OrdersCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "flashcart_orders_created_total",
		Help: "Total number of orders successfully created.",
	})

	// OrdersCancelled counts cancelled orders.
	OrdersCancelled = promauto.NewCounter(prometheus.CounterOpts{
		Name: "flashcart_orders_cancelled_total",
		Help: "Total number of orders cancelled by customers.",
	})

	// OrderTotalAmount tracks revenue distribution per order checkout.
	OrderTotalAmount = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "flashcart_order_total_amount",
		Help:    "Distribution of order total amounts in USD.",
		Buckets: []float64{10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
	})
)

// ── Payment metrics ───────────────────────────────────────────────────────────

var (
	// PaymentsTotal counts payment attempts by final status (captured|failed).
	PaymentsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flashcart_payments_total",
			Help: "Total payment attempts partitioned by final status (captured, failed).",
		},
		[]string{"status"},
	)

	// PaymentAmount observes the distribution of successfully captured payment amounts.
	PaymentAmount = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "flashcart_payment_amount",
		Help:    "Distribution of successfully captured payment amounts in USD.",
		Buckets: []float64{10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
	})
)

// ── Coupon metrics ────────────────────────────────────────────────────────────

var (
	// CouponsValidated counts coupon validation attempts by result.
	CouponsValidated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flashcart_coupons_validated_total",
			Help: "Coupon validation attempts partitioned by result (valid, expired, exhausted, already_used, not_found, inactive).",
		},
		[]string{"result"},
	)
)

// ── Inventory metrics ─────────────────────────────────────────────────────────

var (
	// InventoryLockConflicts counts optimistic lock conflicts during checkout.
	InventoryLockConflicts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "flashcart_inventory_conflicts_total",
		Help: "Total number of optimistic concurrency lock conflicts during stock reservation.",
	})
)

// ── Cart metrics ──────────────────────────────────────────────────────────────

var (
	// CartItemsAdded counts items added to shopping carts.
	CartItemsAdded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "flashcart_cart_items_added_total",
		Help: "Total number of items added to shopping carts.",
	})
)

// ── Worker pool metrics ───────────────────────────────────────────────────────

var (
	// WorkerJobsDispatched counts jobs submitted to the async worker pool by job type.
	WorkerJobsDispatched = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flashcart_worker_jobs_dispatched_total",
			Help: "Total number of background jobs dispatched to the worker pool, by job type.",
		},
		[]string{"job_type"},
	)

	// WorkerJobsProcessed counts jobs successfully completed by workers, by job type.
	WorkerJobsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flashcart_worker_jobs_processed_total",
			Help: "Total number of background jobs completed by the worker pool, by job type.",
		},
		[]string{"job_type"},
	)

	// WorkerJobsFailed counts jobs that exhausted all retries and went to DLQ.
	WorkerJobsFailed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flashcart_worker_jobs_failed_total",
			Help: "Total number of background jobs that failed all retries and were sent to DLQ.",
		},
		[]string{"job_type"},
	)
)

// ── Rate limiting metrics ─────────────────────────────────────────────────────

var (
	// RateLimitHits counts requests blocked by the rate limiter, by route.
	RateLimitHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flashcart_rate_limit_hits_total",
			Help: "Total number of requests rejected by the rate limiter, by route.",
		},
		[]string{"route"},
	)
)

package worker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/logger"
)

// OrderCreatedJob dispatches async post-checkout processing:
// email confirmation, invoice generation trigger, and analytics ingestion.
type OrderCreatedJob struct {
	OrderID   uuid.UUID
	UserID    uuid.UUID
	Total     float64
	CreatedAt time.Time
}

func (j *OrderCreatedJob) Name() string { return "order_created" }

func (j *OrderCreatedJob) Execute(ctx context.Context) {
	// Simulate email notification dispatch (real impl: SMTP / SES / SendGrid)
	logger.Info(ctx, "dispatching order confirmation email",
		"order_id", j.OrderID.String(),
		"user_id", j.UserID.String(),
		"total", j.Total,
	)

	// Simulate analytics event push (real impl: Segment / BigQuery / Kafka)
	logger.Info(ctx, "pushing order analytics event",
		"order_id", j.OrderID.String(),
		"total_usd", j.Total,
		"created_at", j.CreatedAt.Format(time.RFC3339),
	)
}

// InvoiceJob generates and stores a PDF invoice for a completed order.
type InvoiceJob struct {
	OrderID uuid.UUID
	UserID  uuid.UUID
	Total   float64
}

func (j *InvoiceJob) Name() string { return "invoice_generate" }

func (j *InvoiceJob) Execute(ctx context.Context) {
	// Simulate PDF invoice rendering (real impl: wkhtmltopdf / Gotenberg / LaTeX)
	logger.Info(ctx, "generating PDF invoice for order",
		"order_id", j.OrderID.String(),
		"user_id", j.UserID.String(),
		"total", j.Total,
	)

	// Simulate S3 / GCS upload
	logger.Info(ctx, "invoice upload complete",
		"order_id", j.OrderID.String(),
		"destination", "s3://flashcart-invoices/"+j.OrderID.String()+".pdf",
	)
}

// PaymentProcessedJob records a payment confirmation event for analytics / audit.
type PaymentProcessedJob struct {
	PaymentID uuid.UUID
	OrderID   uuid.UUID
	UserID    uuid.UUID
	Amount    float64
	Status    string
}

func (j *PaymentProcessedJob) Name() string { return "payment_processed" }

func (j *PaymentProcessedJob) Execute(ctx context.Context) {
	logger.Info(ctx, "payment event recorded",
		"payment_id", j.PaymentID.String(),
		"order_id", j.OrderID.String(),
		"amount", j.Amount,
		"status", j.Status,
	)
}

// InventoryRestoreJob releases reserved stock when an order is cancelled.
// In production this would call the inventory repository directly.
type InventoryRestoreJob struct {
	ProductID uuid.UUID
	Quantity  int
}

func (j *InventoryRestoreJob) Name() string { return "inventory_restore" }

func (j *InventoryRestoreJob) Execute(ctx context.Context) {
	// Simulate restoring inventory (real impl: call inventoryRepo.ReleaseStock)
	logger.Info(ctx, "inventory restore triggered",
		"product_id", j.ProductID.String(),
		"quantity", j.Quantity,
	)
}

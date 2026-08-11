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

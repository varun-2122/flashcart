package payment

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/auth"
	"github.com/varun-2122/flashcart/internal/domain"
	"github.com/varun-2122/flashcart/internal/response"
)

// PaymentHandler exposes payment endpoints.
type PaymentHandler struct {
	service *PaymentService
}

// NewPaymentHandler creates a PaymentHandler.
func NewPaymentHandler(service *PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

// InitiatePayment handles POST /api/v1/payments
//
// Body: { "order_id": "...", "idempotency_key": "..." (optional) }
func (h *PaymentHandler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	var req struct {
		OrderID        string `json:"order_id"`
		IdempotencyKey string `json:"idempotency_key"`
		Provider       string `json:"provider"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", err.Error())
		return
	}

	orderID, err := uuid.Parse(req.OrderID)
	if err != nil {
		response.BadRequest(w, "Invalid order_id UUID", nil)
		return
	}

	chargeReq := ChargeRequest{
		OrderID:        orderID,
		UserID:         userID,
		IdempotencyKey: req.IdempotencyKey,
		Provider:       req.Provider,
	}

	payment, err := h.service.Charge(r.Context(), chargeReq)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentFailed) {
			response.ErrorJSON(w, http.StatusPaymentRequired, "PAYMENT_FAILED", payment.FailureReason, nil)
			return
		}
		response.InternalServerError(w, err.Error())
		return
	}

	response.Created(w, payment)
}

// GetPayment handles GET /api/v1/payments/{id}
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid payment UUID", nil)
		return
	}

	p, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			response.NotFound(w, "Payment not found")
			return
		}
		response.InternalServerError(w, err.Error())
		return
	}

	response.Success(w, p)
}

// GetOrderPayment handles GET /api/v1/orders/{id}/payment
func (h *PaymentHandler) GetOrderPayment(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	orderIDStr := r.PathValue("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid order UUID", nil)
		return
	}

	p, err := h.service.GetByOrderID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			response.NotFound(w, "No payment found for this order")
			return
		}
		response.InternalServerError(w, err.Error())
		return
	}

	response.Success(w, p)
}

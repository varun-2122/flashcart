package order

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/auth"
	"github.com/varun-2122/flashcart/internal/domain"
	"github.com/varun-2122/flashcart/internal/response"
)

type OrderHandler struct {
	service *OrderService
}

func NewOrderHandler(service *OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	// Optional body: { "coupon_code": "SUMMER20" }
	var req struct {
		CouponCode     string  `json:"coupon_code"`
		CouponDiscount float64 `json:"coupon_discount"` // pre-validated from /coupons/validate
	}
	// Ignore decode errors — body is optional
	_ = json.NewDecoder(r.Body).Decode(&req)

	order, err := h.service.CreateOrderFromCart(r.Context(), userID, req.CouponCode, req.CouponDiscount)
	if err != nil {
		if errors.Is(err, domain.ErrEmptyCart) {
			response.BadRequest(w, err.Error(), nil)
			return
		}
		if errors.Is(err, domain.ErrInsufficientStock) || errors.Is(err, domain.ErrOptimisticLockConflict) {
			response.Conflict(w, err.Error())
			return
		}
		response.InternalServerError(w, err.Error())
		return
	}

	response.Created(w, order)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid order UUID", nil)
		return
	}

	order, err := h.service.GetOrderByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			response.NotFound(w, "Order not found")
			return
		}
		response.InternalServerError(w, err.Error())
		return
	}

	response.Success(w, order)
}

func (h *OrderHandler) ListUserOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	orders, err := h.service.ListUserOrders(r.Context(), userID)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}

	response.Success(w, orders)
}

// CancelOrder handles POST /api/v1/orders/{id}/cancel
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	idStr := r.PathValue("id")
	orderID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid order UUID", nil)
		return
	}

	order, err := h.service.CancelOrder(r.Context(), orderID, userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderNotFound):
			response.NotFound(w, "Order not found")
		case errors.Is(err, domain.ErrOrderNotOwned):
			response.Forbidden(w, err.Error())
		case errors.Is(err, domain.ErrOrderCannotBeCancelled):
			response.Conflict(w, err.Error())
		default:
			response.InternalServerError(w, err.Error())
		}
		return
	}

	response.Success(w, order)
}

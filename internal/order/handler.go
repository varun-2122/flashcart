package order

import (
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
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	order, err := h.service.CreateOrderFromCart(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrEmptyCart) {
			response.BadRequest(w, err.Error(), nil)
			return
		}
		if errors.Is(err, domain.ErrInsufficientStock) || errors.Is(err, domain.ErrOptimisticLockConflict) {
			response.ErrorJSON(w, http.StatusConflict, "STOCK_CONFLICT", err.Error(), nil)
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
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	orders, err := h.service.ListUserOrders(r.Context(), userID)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}

	response.Success(w, orders)
}

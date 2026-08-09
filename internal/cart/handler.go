package cart

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/auth"
	"github.com/varun-2122/flashcart/internal/response"
)

type CartHandler struct {
	service *CartService
}

func NewCartHandler(service *CartService) *CartHandler {
	return &CartHandler{service: service}
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	var req AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", err.Error())
		return
	}

	cart, err := h.service.AddItem(r.Context(), userID, req)
	if err != nil {
		response.BadRequest(w, err.Error(), nil)
		return
	}

	response.Success(w, cart)
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	cart, err := h.service.GetCart(r.Context(), userID)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}

	response.Success(w, cart)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated", nil)
		return
	}

	prodIDStr := r.PathValue("product_id")
	if prodIDStr == "" {
		prodIDStr = r.URL.Query().Get("product_id")
	}

	productID, err := uuid.Parse(prodIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid product_id", nil)
		return
	}

	cart, err := h.service.RemoveItem(r.Context(), userID, productID)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}

	response.Success(w, cart)
}

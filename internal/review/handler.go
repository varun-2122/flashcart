package review

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/auth"
	"github.com/varun-2122/flashcart/internal/domain"
	"github.com/varun-2122/flashcart/internal/response"
)

type ReviewHandler struct {
	service *ReviewService
}

func NewReviewHandler(service *ReviewService) *ReviewHandler {
	return &ReviewHandler{service: service}
}

func (h *ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	productIDStr := r.PathValue("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid product ID format", nil)
		return
	}

	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	review, err := h.service.CreateReview(r.Context(), productID, userID, req.Rating, req.Comment)
	if err != nil {
		if err == domain.ErrInvalidRating {
			response.Error(w, http.StatusBadRequest, err.Error(), nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to create review", err)
		return
	}

	response.Success(w, review)
}

func (h *ReviewHandler) GetReviews(w http.ResponseWriter, r *http.Request) {
	productIDStr := r.PathValue("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid product ID format", nil)
		return
	}

	reviews, err := h.service.GetProductReviews(r.Context(), productID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch reviews", err)
		return
	}

	response.Success(w, reviews)
}

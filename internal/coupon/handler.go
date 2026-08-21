package coupon

import (
	"encoding/json"
	"net/http"

	"github.com/varun-2122/flashcart/internal/auth"
	"github.com/varun-2122/flashcart/internal/response"
)

// CouponHandler exposes coupon validation endpoints.
type CouponHandler struct {
	service *CouponService
}

// NewCouponHandler creates a CouponHandler.
func NewCouponHandler(service *CouponService) *CouponHandler {
	return &CouponHandler{service: service}
}

// ValidateCoupon handles POST /api/v1/coupons/validate
//
// Body: { "code": "SUMMER20" }
// Response: { valid: bool, discount_percent: float, message: string }
func (h *CouponHandler) ValidateCoupon(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", err.Error())
		return
	}

	if req.Code == "" {
		response.BadRequest(w, "coupon code is required", nil)
		return
	}

	result, err := h.service.Validate(r.Context(), req.Code, userID)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}

	response.Success(w, result)
}

package user

import (
	"encoding/json"
	"net/http"

	"github.com/varun-2122/flashcart/internal/auth"
	"github.com/varun-2122/flashcart/internal/response"
)

// UserHandler exposes user profile endpoints.
type UserHandler struct {
	service *UserService
}

// NewUserHandler creates a UserHandler.
func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{service: service}
}

// GetMe handles GET /api/v1/users/me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	profile, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	// Never send password hash
	profile.PasswordHash = ""
	response.Success(w, profile)
}

// UpdateMe handles PATCH /api/v1/users/me
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", err.Error())
		return
	}

	if req.FirstName == "" && req.LastName == "" {
		response.BadRequest(w, "At least one of first_name or last_name must be provided", nil)
		return
	}

	updated, err := h.service.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}

	updated.PasswordHash = ""
	response.Success(w, updated)
}

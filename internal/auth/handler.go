package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/varun-2122/flashcart/internal/domain"
	"github.com/varun-2122/flashcart/internal/response"
)

type AuthHandler struct {
	authService *AuthService
}

func NewAuthHandler(authService *AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", err.Error())
		return
	}

	if req.Email == "" || req.Password == "" || req.FirstName == "" {
		response.BadRequest(w, "email, password, and first_name are required", nil)
		return
	}

	res, err := h.authService.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			response.ErrorJSON(w, http.StatusConflict, "USER_EXISTS", err.Error(), nil)
			return
		}
		response.InternalServerError(w, err.Error())
		return
	}

	response.Created(w, res)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", err.Error())
		return
	}

	if req.Email == "" || req.Password == "" {
		response.BadRequest(w, "email and password are required", nil)
		return
	}

	res, err := h.authService.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			response.ErrorJSON(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error(), nil)
			return
		}
		response.InternalServerError(w, err.Error())
		return
	}

	response.Success(w, res)
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Credential string `json:"credential"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", err.Error())
		return
	}

	if req.Credential == "" {
		response.BadRequest(w, "credential is required", nil)
		return
	}

	res, err := h.authService.GoogleLogin(r.Context(), req.Credential)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}

	response.Success(w, res)
}

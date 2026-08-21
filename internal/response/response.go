package response

import (
	"encoding/json"
	"net/http"
)

// Response represents standardized API response payload.
type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

// Error represents standardized API error payload.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// JSON sends a JSON response with status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := Response{
		Success: status >= 200 && status < 300,
		Data:    data,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// Success sends a standard successful JSON response with 200 OK.
func Success(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, data)
}

// Created sends a standard 201 Created JSON response.
func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, data)
}

// NoContent sends a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Pagination sends a successful response with meta pagination data.
func Pagination(w http.ResponseWriter, data any, meta any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// ErrorJSON sends a structured JSON error response.
func ErrorJSON(w http.ResponseWriter, status int, code, message string, details any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := Response{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// BadRequest sends HTTP 400 Bad Request error.
func BadRequest(w http.ResponseWriter, message string, details any) {
	ErrorJSON(w, http.StatusBadRequest, "BAD_REQUEST", message, details)
}

// Unauthorized sends HTTP 401 Unauthorized error.
func Unauthorized(w http.ResponseWriter, message string) {
	ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

// Forbidden sends HTTP 403 Forbidden error.
func Forbidden(w http.ResponseWriter, message string) {
	ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", message, nil)
}

// NotFound sends HTTP 404 Not Found error.
func NotFound(w http.ResponseWriter, message string) {
	ErrorJSON(w, http.StatusNotFound, "NOT_FOUND", message, nil)
}

// Conflict sends HTTP 409 Conflict error.
func Conflict(w http.ResponseWriter, message string) {
	ErrorJSON(w, http.StatusConflict, "CONFLICT", message, nil)
}

// TooManyRequests sends HTTP 429 Too Many Requests error.
func TooManyRequests(w http.ResponseWriter, message string) {
	ErrorJSON(w, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", message, nil)
}

// InternalServerError sends HTTP 500 Internal Server Error.
func InternalServerError(w http.ResponseWriter, message string) {
	ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message, nil)
}

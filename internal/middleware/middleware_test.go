package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := RequestID(nextHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	reqID := w.Header().Get("X-Request-ID")
	if reqID == "" {
		t.Errorf("expected X-Request-ID header to be populated")
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("simulated critical crash")
	})

	handler := Recovery(panicHandler)

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 on panic recovery, got %d", w.Code)
	}
}

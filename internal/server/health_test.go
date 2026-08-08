package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoints(t *testing.T) {
	handler := NewHealthHandler(nil, nil)

	t.Run("Healthz Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/healthz", nil)
		w := httptest.NewRecorder()

		handler.Healthz(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var body map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &body)

		if !body["success"].(bool) {
			t.Errorf("expected success true")
		}
	})

	t.Run("Livez Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/livez", nil)
		w := httptest.NewRecorder()

		handler.Livez(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Readyz Endpoint - Degraded", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/readyz", nil)
		w := httptest.NewRecorder()

		handler.Readyz(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503 for uninitialized DB/Redis, got %d", w.Code)
		}
	})
}

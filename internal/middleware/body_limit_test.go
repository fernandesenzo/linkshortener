package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fernandesenzo/linkshortener/internal/middleware"
)

func TestBodyLimit(t *testing.T) {
	handler := middleware.BodyLimit(10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("under limit", func(t *testing.T) {
		body := strings.NewReader("1234567890")
		req := httptest.NewRequest("POST", "/", body)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("over limit", func(t *testing.T) {
		body := strings.NewReader("12345678901")
		req := httptest.NewRequest("POST", "/", body)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("expected 413, got %d", rec.Code)
		}
	})

	t.Run("exactly limit", func(t *testing.T) {
		body := strings.NewReader("1234567890")
		req := httptest.NewRequest("POST", "/", body)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

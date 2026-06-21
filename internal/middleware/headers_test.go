package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fernandesenzo/linkshortener/internal/middleware"
)

func TestApplyHeaders(t *testing.T) {
	expectedHeaders := []struct {
		header string
		want   string
	}{
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"Strict-Transport-Security", "max-age=31536000; includeSubDomains"},
		{"Access-Control-Allow-Origin", "*"},
		{"Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE"},
		{"Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID"},
	}

	t.Run("non-OPTIONS request should apply headers and call next handler", func(t *testing.T) {
		var called bool
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		m := middleware.ApplyHeaders(innerHandler)
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		m.ServeHTTP(rec, req)

		if !called {
			t.Error("inner handler was not called")
		}

		for _, tt := range expectedHeaders {
			got := rec.Header().Get(tt.header)
			if got != tt.want {
				t.Errorf("header %q = %q; want %q", tt.header, got, tt.want)
			}
		}

		if rec.Code != http.StatusOK {
			t.Errorf("status code = %d; want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("OPTIONS request should apply headers, return OK and NOT call next handler", func(t *testing.T) {
		var called bool
		innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		m := middleware.ApplyHeaders(innerHandler)
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		rec := httptest.NewRecorder()

		m.ServeHTTP(rec, req)

		if called {
			t.Error("inner handler was called for OPTIONS request")
		}

		for _, tt := range expectedHeaders {
			got := rec.Header().Get(tt.header)
			if got != tt.want {
				t.Errorf("header %q = %q; want %q", tt.header, got, tt.want)
			}
		}

		if rec.Code != http.StatusOK {
			t.Errorf("status code = %d; want %d", rec.Code, http.StatusOK)
		}
	})
}

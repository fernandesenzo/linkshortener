package handler

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestHandler_Get(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		redisServer := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
		t.Cleanup(func() {
			if err := redisClient.Close(); err != nil {
				t.Errorf("close redis client: %v", err)
			}
		})

		h := New(redisClient)
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if got := rec.Header().Get("Content-Type"); got != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", got)
		}
		if got, want := rec.Body.String(), "{\"message\":\"ok\"}\n"; got != want {
			t.Errorf("expected body %q, got %q", want, got)
		}
	})

	t.Run("redis unavailable", func(t *testing.T) {
		redisClient := redis.NewClient(&redis.Options{
			MaxRetries: -1,
			Dialer: func(context.Context, string, string) (net.Conn, error) {
				return nil, errors.New("redis unavailable")
			},
		})
		t.Cleanup(func() {
			if err := redisClient.Close(); err != nil {
				t.Errorf("close redis client: %v", err)
			}
		})

		h := New(redisClient)
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		h.Get(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
		}
		if got, want := rec.Body.String(), "error\n"; got != want {
			t.Errorf("expected body %q, got %q", want, got)
		}
	})
}

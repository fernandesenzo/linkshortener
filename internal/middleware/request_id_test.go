package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fernandesenzo/linkshortener/internal/logger"
	"github.com/fernandesenzo/linkshortener/internal/middleware"
	"github.com/google/uuid"
)

func TestInjectReqID(t *testing.T) {
	tests := []struct {
		name         string
		headerReqID  string
		wantCustomID bool
	}{
		{
			name:         "uses existing X-Request-ID header",
			headerReqID:  "existing-req-id-12345",
			wantCustomID: true,
		},
		{
			name:         "generates new UUID request ID if header is missing",
			headerReqID:  "",
			wantCustomID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called bool
			var receivedID string

			innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				if id, ok := r.Context().Value(logger.RequestIDKey).(string); ok {
					receivedID = id
				}
			})

			m := middleware.InjectReqID(innerHandler)
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.headerReqID != "" {
				req.Header.Set("X-Request-ID", tt.headerReqID)
			}

			rec := httptest.NewRecorder()

			m.ServeHTTP(rec, req)

			if !called {
				t.Error("inner handler was not called")
			}

			if tt.wantCustomID {
				if receivedID != tt.headerReqID {
					t.Errorf("got request ID = %q, want %q", receivedID, tt.headerReqID)
				}
			} else {
				if receivedID == "" {
					t.Error("expected a generated request ID, but got empty string")
				} else {
					if _, err := uuid.Parse(receivedID); err != nil {
						t.Errorf("expected a valid UUID request ID, but got error: %v (raw: %q)", err, receivedID)
					}
				}
			}
		})
	}
}

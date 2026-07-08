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
		headerCFRay  string
		wantCustomID bool
		wantID       string
	}{
		{
			name:         "uses existing X-Request-ID header",
			headerReqID:  "existing-req-id-12345",
			wantCustomID: true,
			wantID:       "existing-req-id-12345",
		},
		{
			name:         "uses CF-Ray header if X-Request-ID is missing",
			headerCFRay:  "5cf-ray-id-67890",
			wantCustomID: true,
			wantID:       "5cf-ray-id-67890",
		},
		{
			name:         "prefers X-Request-ID header over CF-Ray header",
			headerReqID:  "existing-req-id-12345",
			headerCFRay:  "5cf-ray-id-67890",
			wantCustomID: true,
			wantID:       "existing-req-id-12345",
		},
		{
			name:         "generates new UUID request ID if headers are missing",
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
			if tt.headerCFRay != "" {
				req.Header.Set("CF-Ray", tt.headerCFRay)
			}

			rec := httptest.NewRecorder()

			m.ServeHTTP(rec, req)

			if !called {
				t.Error("inner handler was not called")
			}

			if tt.wantCustomID {
				if receivedID != tt.wantID {
					t.Errorf("got request ID = %q, want %q", receivedID, tt.wantID)
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

			respHeader := rec.Header().Get("X-Request-ID")
			if respHeader != receivedID {
				t.Errorf("got response header X-Request-ID = %q, want %q", respHeader, receivedID)
			}
		})
	}
}

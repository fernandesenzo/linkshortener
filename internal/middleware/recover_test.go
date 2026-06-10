package middleware_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fernandesenzo/linkshortener/internal/middleware"
)

func TestRecover(t *testing.T) {
	tests := []struct {
		name          string
		shouldPanic   bool
		panicVal      interface{}
		wantStatus    int
		wantBody      string
		expectLogErr  string
	}{
		{
			name:        "no panic",
			shouldPanic: false,
			wantStatus:  http.StatusOK,
			wantBody:    "ok",
		},
		{
			name:         "panic with string",
			shouldPanic:  true,
			panicVal:     "something went wrong",
			wantStatus:   http.StatusInternalServerError,
			wantBody:     "internal server error\n",
			expectLogErr: "something went wrong",
		},
		{
			name:         "panic with error",
			shouldPanic:  true,
			panicVal:     errors.New("database connection failed"),
			wantStatus:   http.StatusInternalServerError,
			wantBody:     "internal server error\n",
			expectLogErr: "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			testHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError})
			testLogger := slog.New(testHandler)

			oldLogger := slog.Default()
			slog.SetDefault(testLogger)
			defer slog.SetDefault(oldLogger)

			innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.shouldPanic {
					panic(tt.panicVal)
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			})

			m := middleware.Recover(innerHandler)

			req := httptest.NewRequest("GET", "/test", nil)
			rec := httptest.NewRecorder()

			m.ServeHTTP(rec, req)

			// Verify response status
			if rec.Code != tt.wantStatus {
				t.Errorf("got status = %d, want %d", rec.Code, tt.wantStatus)
			}

			// Verify response body
			if rec.Body.String() != tt.wantBody {
				t.Errorf("got body = %q, want %q", rec.Body.String(), tt.wantBody)
			}

			// Verify logs if panic happened
			if tt.shouldPanic {
				var parsed map[string]interface{}
				err := json.Unmarshal(buf.Bytes(), &parsed)
				if err != nil {
					t.Fatalf("failed to parse log JSON: %v, output: %s", err, buf.String())
				}

				if parsed["msg"] != "request panicked" {
					t.Errorf("got msg = %v, want 'request panicked'", parsed["msg"])
				}

				errField := parsed["err"]
				if errField == nil {
					t.Error("missing 'err' field in logs")
				} else {
					errStr, ok := errField.(string)
					if !ok {
						t.Errorf("expected 'err' field to be string, got %T (%v)", errField, errField)
					} else if !strings.Contains(errStr, tt.expectLogErr) {
						t.Errorf("got log err = %q, want it to contain %q", errStr, tt.expectLogErr)
					}
				}
			} else {
				if buf.Len() > 0 {
					t.Errorf("expected no logs on successful request, but got: %s", buf.String())
				}
			}
		})
	}
}

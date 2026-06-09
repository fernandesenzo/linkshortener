package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fernandesenzo/linkshortener/internal/middleware"
)

func TestAccessLog(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		path          string
		remoteAddr    string
		handlerStatus int
		latencyDelay  time.Duration
	}{
		{
			name:          "successful get request",
			method:        "GET",
			path:          "/info",
			remoteAddr:    "192.168.1.1:1234",
			handlerStatus: http.StatusOK,
			latencyDelay:  10 * time.Millisecond,
		},
		{
			name:          "failed post request",
			method:        "POST",
			path:          "/create",
			remoteAddr:    "10.0.0.1:5678",
			handlerStatus: http.StatusBadRequest,
			latencyDelay:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			testHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
			testLogger := slog.New(testHandler)

			oldLogger := slog.Default()
			slog.SetDefault(testLogger)
			defer slog.SetDefault(oldLogger)

			innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.latencyDelay > 0 {
					time.Sleep(tt.latencyDelay)
				}
				w.WriteHeader(tt.handlerStatus)
			})

			m := middleware.AccessLog(innerHandler)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.RemoteAddr = tt.remoteAddr
			rec := httptest.NewRecorder()

			m.ServeHTTP(rec, req)

			var parsed map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &parsed)
			if err != nil {
				t.Fatalf("failed to parse log JSON: %v, output: %s", err, buf.String())
			}

			if parsed["msg"] != "request finished" {
				t.Errorf("got msg = %v, want 'request finished'", parsed["msg"])
			}

			if parsed["method"] != tt.method {
				t.Errorf("got method = %v, want %v", parsed["method"], tt.method)
			}

			if parsed["path"] != tt.path {
				t.Errorf("got path = %v, want %v", parsed["path"], tt.path)
			}

			if parsed["ip"] != tt.remoteAddr {
				t.Errorf("got ip = %v, want %v", parsed["ip"], tt.remoteAddr)
			}

			statusFloat, ok := parsed["status"].(float64)
			if !ok {
				t.Errorf("expected status to be a number, got %T", parsed["status"])
			} else if int(statusFloat) != tt.handlerStatus {
				t.Errorf("got status = %v, want %v", statusFloat, tt.handlerStatus)
			}

			latencyVal, ok := parsed["latency_ms"].(float64)
			if !ok {
				t.Errorf("expected latency_ms to be a number, got %T", parsed["latency_ms"])
			} else if int64(latencyVal) < 0 {
				t.Errorf("got latency_ms = %v, want >= 0", latencyVal)
			}
		})
	}
}

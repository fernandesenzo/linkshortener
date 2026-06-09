package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestCustomHandler(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		msg       string
		wantReqID string
		hasReqID  bool
	}{
		{
			name:      "with request id",
			ctx:       context.WithValue(context.Background(), RequestIDKey, "req-12345"),
			msg:       "test message with id",
			wantReqID: "req-12345",
			hasReqID:  true,
		},
		{
			name:      "without request id",
			ctx:       context.Background(),
			msg:       "test message without id",
			wantReqID: "",
			hasReqID:  false,
		},
		{
			name:      "with non-string request id",
			ctx:       context.WithValue(context.Background(), RequestIDKey, 12345),
			msg:       "test message with invalid id type",
			wantReqID: "",
			hasReqID:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			baseHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			})

			h := &customHandler{Handler: baseHandler}
			l := slog.New(h)

			l.InfoContext(tt.ctx, tt.msg)

			var parsed map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &parsed)
			if err != nil {
				t.Fatalf("failed to parse log output: %v, raw output: %s", err, buf.String())
			}

			if parsed["msg"] != tt.msg {
				t.Errorf("got msg = %v, want %v", parsed["msg"], tt.msg)
			}

			val, exists := parsed["req_id"]
			if tt.hasReqID {
				if !exists {
					t.Errorf("expected 'req_id' key in log, but it was missing; raw: %s", buf.String())
				} else if val != tt.wantReqID {
					t.Errorf("got req_id = %v, want %v", val, tt.wantReqID)
				}
			} else {
				if exists {
					t.Errorf("did not expect 'req_id' key in log, but got: %v; raw: %s", val, buf.String())
				}
			}
		})
	}
}

func TestSetup(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Setup() panicked: %v", r)
		}
	}()
	Setup()
}

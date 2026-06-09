package logger

import (
	"context"
	"log/slog"
	"os"
)

type contextKey string

const RequestIDKey contextKey = "reqID"

type customHandler struct {
	slog.Handler
}

func (h *customHandler) Handle(ctx context.Context, r slog.Record) error {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		r.AddAttrs(slog.String("req_id", reqID))
	}
	return h.Handler.Handle(ctx, r)
}

func Setup() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, opts)

	handlerWithAttr := jsonHandler.WithAttrs([]slog.Attr{
		slog.String("app", "shortener"),
	})

	finalHandler := customHandler{Handler: handlerWithAttr}

	logger := slog.New(&finalHandler)

	slog.SetDefault(logger)
}

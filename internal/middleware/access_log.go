package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriterObserver struct {
	http.ResponseWriter
	status int
}

func (w *responseWriterObserver) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		observer := &responseWriterObserver{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(observer, r)

		slog.InfoContext(r.Context(), "request finished",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("ip", r.RemoteAddr),
			slog.Int("status", observer.status),
			slog.Int64("latency_ms", time.Since(start).Milliseconds()),
		)
	})
}

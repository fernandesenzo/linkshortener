package middleware

import (
	"context"
	"net/http"

	"github.com/fernandesenzo/linkshortener/internal/logger"
	"github.com/google/uuid"
)

func InjectReqID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// might come from cf
		uid := r.Header.Get("X-Request-ID")
		if uid == "" {
			uid = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), logger.RequestIDKey, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

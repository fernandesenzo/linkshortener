package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/redis/go-redis/v9"
)

type Handler struct {
	client *redis.Client
}

func New(client *redis.Client) *Handler {
	return &Handler{client: client}
}

type healthResponse struct {
	Message string `json:"message"`
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {

	if err := h.client.Ping(r.Context()).Err(); err != nil {
		slog.ErrorContext(r.Context(), "health.handler.Get: failed to ping redis:", "err", err.Error())
		http.Error(w, "error", http.StatusServiceUnavailable)
		return
	}
	resp := healthResponse{Message: "ok"}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.ErrorContext(r.Context(), "health.handler.Get: failed to encode response", "err", err.Error())
	}

}

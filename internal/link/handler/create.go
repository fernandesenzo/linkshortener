package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/fernandesenzo/linkshortener/internal/link"
)

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		http.Error(w, "unsupported content type", http.StatusUnsupportedMediaType)
		return
	}

	var req createLinkRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	//TODO: when moving to cloudflare, use forwarded for instead
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to obtain ip from remote address", "ip", r.RemoteAddr)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	l, err := h.svc.CreateLink(r.Context(), ip, req.URL)
	if err != nil {
		handleCreateError(err, w, r.Context())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := createLinkResponse{
		Code: l.Code,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.ErrorContext(r.Context(), "handler.Create: failed to encode response", "err", err.Error())
	}

}

func handleCreateError(err error, w http.ResponseWriter, ctx context.Context) {
	switch {
	case errors.Is(err, link.ErrInvalidURL):
		http.Error(w, "invalid url", http.StatusBadRequest)

	case errors.Is(err, link.ErrTooLongURL):
		http.Error(w, "url too long", http.StatusUnprocessableEntity)

	case errors.Is(err, link.ErrTooManyActiveURLs):
		http.Error(w, "ip already has the limit of links shortened, try again later.", http.StatusUnprocessableEntity)

	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
		slog.ErrorContext(ctx, "handler.Create: unknown error on create request", "err", err)
	}
}

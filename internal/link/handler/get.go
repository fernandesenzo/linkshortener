package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/fernandesenzo/linkshortener/internal/link"
)

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if len(code) != link.CodeLength {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	l, err := h.svc.GetLink(r.Context(), code)

	if err != nil {
		if errors.Is(err, link.ErrNotFound) {
			http.Error(w, "no link with this code", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		slog.ErrorContext(r.Context(), "handler.Get: error getting link", "err", err, "code", code)
		return
	}

	http.Redirect(w, r, l.OriginalURL, http.StatusTemporaryRedirect)
}

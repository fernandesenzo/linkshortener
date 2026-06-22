package handler

import (
	"context"

	"github.com/fernandesenzo/linkshortener/internal/link"
)

type Handler struct {
	svc Service
}

func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

type Service interface {
	CreateLink(ctx context.Context, ip string, url string) (l *link.Link, err error)
	GetLink(ctx context.Context, code string) (*link.Link, error)
}

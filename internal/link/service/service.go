package service

import (
	"context"

	"github.com/fernandesenzo/linkshortener/internal/link"
)

type CodeGenerator interface {
	Generate(length int) (string, error)
}

type LinkRepository interface {
	CreateIfNotExists(context.Context, *link.Link, string) error
	CountByIPAndIncrement(ctx context.Context, ip string) (int, error)
	GetByCode(ctx context.Context, code string) (*link.Link, error)
	DecrementIPCounter(ctx context.Context, ip string) error
}
type Service struct {
	codeGen  CodeGenerator
	linkRepo LinkRepository
}

func NewService(codeGen CodeGenerator, linkRepo LinkRepository) *Service {
	return &Service{codeGen: codeGen, linkRepo: linkRepo}
}

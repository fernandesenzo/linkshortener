package service

import (
	"context"

	"github.com/fernandesenzo/linkshortener/internal/link"
)

type CodeGenerator interface {
	Generate(length int) (string, error)
}

type LinkRepository interface {
	GetIPLock(ctx context.Context, ip string) (unlock func(), err error)
	Create(ctx context.Context, l *link.Link, ip string) error
	CountByIP(ctx context.Context, ip string) (int, error)
	GetByCode(ctx context.Context, code string) (*link.Link, error)
	IncrementIPCounter(ctx context.Context, ip string) error
}
type Service struct {
	codeGen  CodeGenerator
	linkRepo LinkRepository
}

func New(codeGen CodeGenerator, linkRepo LinkRepository) *Service {
	return &Service{codeGen: codeGen, linkRepo: linkRepo}
}

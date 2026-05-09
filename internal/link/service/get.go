package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/fernandesenzo/linkshortener/internal/link"
	"github.com/fernandesenzo/linkshortener/internal/link/repository"
)

func (s *Service) GetLink(ctx context.Context, code string) (*link.Link, error) {
	l, err := s.linkRepo.GetByCode(ctx, code)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("service.GetLink: failed to get link: %w", err)
		}
		return nil, err
	}
	return l, nil
}

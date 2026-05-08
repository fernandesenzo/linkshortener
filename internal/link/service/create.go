package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/fernandesenzo/linkshortener/internal/link"
	"github.com/fernandesenzo/linkshortener/internal/link/repository"
)

func (s *Service) CreateLink(ctx context.Context, ip string, url string) (l *link.Link, err error) {
	ipCount, err := s.linkRepo.CountByIPAndIncrement(ctx, ip)
	if err != nil {
		return nil, fmt.Errorf("service.CreateLink: failed to count ip usage: %w", err)
	}

	defer func() {
		if err != nil {
			if decrementErr := s.linkRepo.DecrementIPCounter(ctx, ip); decrementErr != nil {
				err = fmt.Errorf("service.CreateLink: failed to decrement ip counter: %w", err)
			}
		}
	}()

	if err = link.CanCreate(url, ipCount-1); err != nil {
		return nil, err
	}

	for i := 0; i < link.CreateLinkMaxAttempts; i++ {
		//avoiding shadowing
		var code string
		code, err = s.codeGen.Generate(link.CodeLength)
		if err != nil {
			err = fmt.Errorf("service.CreateLink: failed to generate code: %w", err)
			return nil, err
		}

		l = &link.Link{
			OriginalURL: url,
			Code:        code,
		}

		if err = s.linkRepo.CreateIfNotExists(ctx, l, ip); err != nil {
			if !errors.Is(err, repository.ErrConflict) {
				return nil, fmt.Errorf("service.CreateLink: failed to create link: %w", err)
			}
			continue
		}
		return l, nil
	}
	return nil, link.ErrTooManyCollisions
}

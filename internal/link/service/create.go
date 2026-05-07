package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/fernandesenzo/linkshortener/internal/link"
	"github.com/fernandesenzo/linkshortener/internal/link/repository"
)

func (s *Service) CreateLink(ctx context.Context, ip string, url string) (*link.Link, error) {
	//TODO: someone mean could send a lot of requests at the same time and they all read the same cached value...
	ipCount, err := s.linkRepo.CountByIP(ctx, ip)
	if err != nil {
		return nil, fmt.Errorf("service.CreateLink: failed to count ip usage: %w", err)
	}

	if err = link.CanCreate(url, ipCount); err != nil {
		return nil, err
	}

	for i := 0; i < link.CreateLinkMaxAttempts; i++ {
		code, err := s.codeGen.Generate(link.CodeLength)
		if err != nil {
			// here, it is intentional to not retry till max attempts are not reached, this limit is only for code collisions.
			return nil, fmt.Errorf("service.CreateLink: failed to generate code: %w", err)
		}

		newLink := link.Link{
			OriginalURL: url,
			Code:        code,
		}

		if err := s.linkRepo.CreateIfNotExists(ctx, &newLink, ip); err != nil {
			if !errors.Is(err, repository.ErrConflict) {
				// same here
				return nil, fmt.Errorf("service.CreateLink: failed to create link: %w", err)
			}
			continue
		}
		return &newLink, nil
	}
	return nil, link.ErrTooManyCollisions
}

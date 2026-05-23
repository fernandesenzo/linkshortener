package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/fernandesenzo/linkshortener/internal/link"
	"github.com/fernandesenzo/linkshortener/internal/link/repository"
)

func (s *Service) CreateLink(ctx context.Context, ip string, url string) (l *link.Link, err error) {
	unlock, err := s.linkRepo.GetIPLock(ctx, ip)
	if err != nil {
		return nil, fmt.Errorf("service.CreateLink: failed to get ip lock: %w", err)
	}
	defer unlock()
	ipCount, err := s.linkRepo.CountByIP(ctx, ip)
	if err != nil {
		return nil, fmt.Errorf("service.CreateLink: failed to count links by ip: %w", err)
	}

	if err = link.CanCreate(url, ipCount); err != nil {
		return nil, err
	}

	for i := 0; i < link.CreateLinkMaxAttempts; i++ {
		code, err := s.codeGen.Generate(link.CodeLength)
		if err != nil {
			return nil, fmt.Errorf("service.CreateLink: failed to generate code: %w", err)
		}
		_, err = s.linkRepo.GetByCode(ctx, code)
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("service.CreateLink: failed to check if code exists: %w", err)
		}
		if err == nil {
			continue
		}
		l = &link.Link{
			OriginalURL: url,
			Code:        code,
		}

		if err := s.linkRepo.Create(ctx, l, ip); err != nil {
			if errors.Is(err, repository.ErrConflict) {
				continue
			}
			return nil, fmt.Errorf("service.CreateLink: failed to create link: %w", err)
		}

		if err := s.linkRepo.IncrementIPCounter(ctx, ip); err != nil {
			//TODO: inject a structured logger into the service
			slog.WarnContext(ctx, "failed to increment ip counter after creating link",
				"ip", ip,
				"code", code,
				"error", err,
			)
		}

		return l, nil
	}
	return nil, link.ErrTooManyCollisions
}

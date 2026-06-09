package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/fernandesenzo/linkshortener/internal/link"
	"github.com/redis/go-redis/v9"
)

func (r *RedisRepository) Create(ctx context.Context, l *link.Link, ip string) error {
	linkKey := linkprefix + l.Code
	ipKey := ipSetPrefix + ip

	ok, err := r.client.SetNX(ctx, linkKey, l.OriginalURL, time.Until(l.ExpiresAt)).Result()

	if err != nil {
		return fmt.Errorf("repository.Create: error setting link: %w", err)
	}
	if !ok {
		return ErrConflict
	}

	_, err = r.client.ZAdd(ctx, ipKey, redis.Z{Score: float64(l.ExpiresAt.Unix()), Member: l.Code}).Result()
	if err != nil {
		if _, err1 := r.client.Del(ctx, linkKey).Result(); err1 != nil {
			slog.WarnContext(ctx, "repository.Create: failed to delete link after not adding it to the ipset",
				"ip", ip,
				"code", l.Code,
				"error", err1,
			)
		}
		return fmt.Errorf("repository.Create: error adding link to ipset: %w", err)
	}

	return nil
}

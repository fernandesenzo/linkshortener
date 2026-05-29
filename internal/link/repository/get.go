package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/fernandesenzo/linkshortener/internal/link"
	"github.com/redis/go-redis/v9"
)

func (r *RedisRepository) GetByCode(ctx context.Context, code string) (*link.Link, error) {
	key := linkprefix + code
	url, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository.GetByCode: failed to get link from redis client: %w", err)
	}
	return &link.Link{
		Code:        code,
		OriginalURL: url,
	}, nil
}

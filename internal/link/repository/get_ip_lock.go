package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-redsync/redsync/v4"
)

func (r *RedisRepository) GetIPLock(ctx context.Context, ip string) (func(), error) {
	mutexName := "shortener:lock:ip:" + ip
	// 2 here is an arbitrary value, gotta think better about what should go there.
	mutex := r.rs.NewMutex(mutexName, redsync.WithExpiry(time.Second*2))

	if err := mutex.LockContext(ctx); err != nil {
		return nil, fmt.Errorf("repository.GetIPLock: failed to acquire mutex: %w", err)
	}
	return func() {
		if ok, err := mutex.Unlock(); !ok || err != nil {
			slog.ErrorContext(ctx, "repository.GetIPLock: failed to unlock mutex",
				"ip", ip,
				"error", err)
		}
	}, nil
}

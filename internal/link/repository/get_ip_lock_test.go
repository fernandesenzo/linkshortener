package repository

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisRepository_GetIPLock(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	repo := NewRedisRepository(client)
	ctx := context.Background()

	t.Run("acquire lock successfully and release it", func(t *testing.T) {
		ip := "192.168.1.1"

		// 1. Acquire lock
		unlock, err := repo.GetIPLock(ctx, ip)
		if err != nil {
			t.Fatalf("unexpected error acquiring lock: %v", err)
		}
		if unlock == nil {
			t.Fatal("expected unlock function to be returned, got nil")
		}

		// 2. Try to acquire lock again (should fail)
		_, err = repo.GetIPLock(ctx, ip)
		if err == nil {
			t.Fatal("expected error acquiring already locked mutex, got nil")
		}

		// 3. Release lock
		unlock()

		// 4. Try to acquire lock again (should succeed now)
		unlock2, err := repo.GetIPLock(ctx, ip)
		if err != nil {
			t.Fatalf("unexpected error acquiring lock after release: %v", err)
		}
		if unlock2 == nil {
			t.Fatal("expected unlock function to be returned, got nil")
		}
		unlock2()
	})

	t.Run("lock expires automatically", func(t *testing.T) {
		ip := "192.168.1.2"

		// 1. Acquire lock
		unlock, err := repo.GetIPLock(ctx, ip)
		if err != nil {
			t.Fatalf("unexpected error acquiring lock: %v", err)
		}
		defer unlock()

		// 2. Advance time in miniredis to expire the lock
		s.FastForward(time.Second * 3)

		// 3. Try to acquire lock again (should succeed since it expired)
		unlock2, err := repo.GetIPLock(ctx, ip)
		if err != nil {
			t.Fatalf("expected lock to be acquirable after expiration, got err: %v", err)
		}
		if unlock2 == nil {
			t.Fatal("expected unlock function to be returned, got nil")
		}
		unlock2()
	})
}

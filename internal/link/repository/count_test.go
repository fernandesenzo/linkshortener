package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisRepository_CountByIP(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	repo := NewRedisRepository(client)
	ctx := context.Background()

	t.Run("CountByIP", func(t *testing.T) {
		tests := []struct {
			name        string
			ip          string
			setup       func()
			wantCount   int
			wantErr     bool
			errContains string
		}{
			{
				name:      "no links for IP",
				ip:        "192.168.0.1",
				setup:     func() {},
				wantCount: 0,
				wantErr:   false,
			},
			{
				name: "only active links",
				ip:   "192.168.0.2",
				setup: func() {
					ipKey := ipSetPrefix + "192.168.0.2"
					future := float64(time.Now().Add(time.Hour).Unix())
					_, _ = s.ZAdd(ipKey, future, "code1")
					_, _ = s.ZAdd(ipKey, future+10, "code2")
				},
				wantCount: 2,
				wantErr:   false,
			},
			{
				name: "only expired links",
				ip:   "192.168.0.3",
				setup: func() {
					ipKey := ipSetPrefix + "192.168.0.3"
					past := float64(time.Now().Add(-time.Hour).Unix())
					_, _ = s.ZAdd(ipKey, past, "code1")
					_, _ = s.ZAdd(ipKey, past-10, "code2")
				},
				wantCount: 0,
				wantErr:   false,
			},
			{
				name: "mixed active and expired links",
				ip:   "192.168.0.4",
				setup: func() {
					ipKey := ipSetPrefix + "192.168.0.4"
					future := float64(time.Now().Add(time.Hour).Unix())
					past := float64(time.Now().Add(-time.Hour).Unix())
					_, _ = s.ZAdd(ipKey, future, "active1")
					_, _ = s.ZAdd(ipKey, past, "expired1")
					_, _ = s.ZAdd(ipKey, future+10, "active2")
					_, _ = s.ZAdd(ipKey, past-10, "expired2")
				},
				wantCount: 2,
				wantErr:   false,
			},
			{
				name: "error case - key is not a sorted set",
				ip:   "192.168.0.5",
				setup: func() {
					ipKey := ipSetPrefix + "192.168.0.5"
					_ = s.Set(ipKey, "not-a-zset")
				},
				wantCount:   0,
				wantErr:     true,
				errContains: "WRONGTYPE",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.setup != nil {
					tt.setup()
				}

				count, err := repo.CountByIP(ctx, tt.ip)
				if tt.wantErr {
					if err == nil {
						t.Errorf("CountByIP() expected error, got nil")
					} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
						t.Errorf("CountByIP() error = %v, must contain %q", err, tt.errContains)
					}
				} else {
					if err != nil {
						t.Errorf("CountByIP() unexpected error = %v", err)
					}
					if count != tt.wantCount {
						t.Errorf("CountByIP() count = %d, wantCount = %d", count, tt.wantCount)
					}
				}
			})
		}
	})
}

package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/fernandesenzo/linkshortener/internal/link"
	"github.com/redis/go-redis/v9"
)

func TestRedisRepository_Create(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	repo := NewRedisRepository(client)
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name        string
			link        *link.Link
			ip          string
			setup       func()
			wantErr     error
			errContains string
		}{
			{
				name: "save new temporary link",
				link: &link.Link{
					Code:        "abc",
					OriginalURL: "https://test.com",
					ExpiresAt:   time.Now().Add(time.Hour),
				},
				ip:      "127.0.0.1",
				wantErr: nil,
			},
			{
				name: "conflict - resource already exists",
				link: &link.Link{
					Code:        "abc",
					OriginalURL: "https://another.com",
					ExpiresAt:   time.Now().Add(time.Hour),
				},
				ip:      "127.0.0.1",
				wantErr: ErrConflict,
			},
			{
				name: "error adding to ipset - trigger rollback",
				link: &link.Link{
					Code:        "rollback-test",
					OriginalURL: "https://test.com",
					ExpiresAt:   time.Now().Add(time.Hour),
				},
				ip: "127.0.0.2",
				setup: func() {
					// define the IP key as String to force ZAdd to fail with WRONGTYPE
					_ = s.Set(ipSetPrefix+"127.0.0.2", "not-a-zset")
				},
				errContains: "WRONGTYPE",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.setup != nil {
					tt.setup()
				}

				err := repo.Create(ctx, tt.link, tt.ip)
				if tt.wantErr != nil {
					if !errors.Is(err, tt.wantErr) {
						t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
					}
				} else if tt.errContains != "" {
					if err == nil {
						t.Errorf("Create() expected error containing %q, got nil", tt.errContains)
					} else if !strings.Contains(err.Error(), tt.errContains) {
						t.Errorf("Create() error = %v, must contain %q", err, tt.errContains)
					}
				} else {
					if err != nil {
						t.Errorf("Create() unexpected error = %v", err)
					}

					val, _ := s.Get(linkprefix + tt.link.Code)
					if val != tt.link.OriginalURL {
						t.Errorf("Value mismatch: got %v, want %v", val, tt.link.OriginalURL)
					}

					ttl := s.TTL(linkprefix + tt.link.Code)
					if ttl <= 0 {
						t.Errorf("expected TTL to be set, got %v", ttl)
					}
				}

				if tt.name == "error adding to ipset - trigger rollback" {
					if s.Exists(linkprefix + tt.link.Code) {
						t.Errorf("expected key to be rolled back/deleted from redis")
					}
				}
			})
		}
	})
}

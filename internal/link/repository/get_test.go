package repository

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisRepository_GetByCode(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	repo := NewRedisRepository(client)
	ctx := context.Background()

	t.Run("GetByCode", func(t *testing.T) {
		tests := []struct {
			name        string
			code        string
			setup       func()
			wantURL     string
			wantErr     error
			errContains string
		}{
			{
				name: "success - get existing link",
				code: "xyz",
				setup: func() {
					_ = s.Set(linkprefix+"xyz", "https://destination.com")
				},
				wantURL: "https://destination.com",
				wantErr: nil,
			},
			{
				name:    "not found - code does not exist",
				code:    "nonexistent",
				setup:   func() {},
				wantErr: ErrNotFound,
			},
			{
				name: "error - wrong type in redis key",
				code: "wrongtype-key",
				setup: func() {
					// define as hash to force GET to fail with WRONGTYPE
					s.HSet(linkprefix+"wrongtype-key", "field", "value")
				},
				errContains: "WRONGTYPE",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.setup != nil {
					tt.setup()
				}

				l, err := repo.GetByCode(ctx, tt.code)
				if tt.wantErr != nil {
					if !errors.Is(err, tt.wantErr) {
						t.Errorf("GetByCode() error = %v, wantErr %v", err, tt.wantErr)
					}
				} else if tt.errContains != "" {
					if err == nil {
						t.Errorf("GetByCode() expected error containing %q, got nil", tt.errContains)
					} else if !strings.Contains(err.Error(), tt.errContains) {
						t.Errorf("GetByCode() error = %v, must contain %q", err, tt.errContains)
					}
				} else {
					if err != nil {
						t.Errorf("GetByCode() unexpected error = %v", err)
					}
					if l == nil {
						t.Fatal("GetByCode() returned nil link on success")
					}
					if l.Code != tt.code {
						t.Errorf("GetByCode() code = %v, want %v", l.Code, tt.code)
					}
					if l.OriginalURL != tt.wantURL {
						t.Errorf("GetByCode() OriginalURL = %v, want %v", l.OriginalURL, tt.wantURL)
					}
				}
			})
		}
	})
}

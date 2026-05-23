package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fernandesenzo/linkshortener/internal/link"
	"github.com/fernandesenzo/linkshortener/internal/link/repository"
	"github.com/fernandesenzo/linkshortener/internal/link/service"
)

func TestService_GetLink(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		code      string
		setupRepo func(m *MockRepository)
		wantLink  *link.Link
		wantErr   error
	}{
		{
			name: "success - link found",
			code: "ABCDEF",
			setupRepo: func(m *MockRepository) {
				m.getByCodeFunc = func(ctx context.Context, code string) (*link.Link, error) {
					return &link.Link{
						OriginalURL: "https://example.com",
						Code:        "ABCDEF",
					}, nil
				}
			},
			wantLink: &link.Link{
				OriginalURL: "https://example.com",
				Code:        "ABCDEF",
			},
			wantErr: nil,
		},
		{
			name: "error - link not found",
			code: "NOTFND",
			setupRepo: func(m *MockRepository) {
				m.getByCodeFunc = func(ctx context.Context, code string) (*link.Link, error) {
					return nil, repository.ErrNotFound
				}
			},
			wantLink: nil,
			wantErr:  repository.ErrNotFound,
		},
		{
			name: "error - repository failure",
			code: "DBERR1",
			setupRepo: func(m *MockRepository) {
				m.getByCodeFunc = func(ctx context.Context, code string) (*link.Link, error) {
					return nil, errors.New("database connection lost")
				}
			},
			wantLink: nil,
			wantErr:  errors.New("service.GetLink: failed to get link: database connection lost"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{}
			tt.setupRepo(repo)

			s := service.New(nil, repo)
			l, err := s.GetLink(ctx, tt.code)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if l.Code != tt.wantLink.Code {
					t.Errorf("expected code %s, got %s", tt.wantLink.Code, l.Code)
				}
				if l.OriginalURL != tt.wantLink.OriginalURL {
					t.Errorf("expected URL %s, got %s", tt.wantLink.OriginalURL, l.OriginalURL)
				}
			}
		})
	}
}

package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fernandesenzo/linkshortener/internal/link"
	"github.com/fernandesenzo/linkshortener/internal/link/repository"
	"github.com/fernandesenzo/linkshortener/internal/link/service"
)

type MockCodeGenerator struct {
	codes []string
	err   error
	index int
}

func (m *MockCodeGenerator) Generate(len int) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	code := m.codes[m.index]
	m.index++
	return code, nil
}

type MockRepository struct {
	countByIPFunc     func(ip string) (int, error)
	createIfNotExists func(l *link.Link, ip string) error
}

func (m *MockRepository) CountByIP(_ context.Context, ip string) (int, error) {
	return m.countByIPFunc(ip)
}

func (m *MockRepository) CreateIfNotExists(_ context.Context, l *link.Link, ip string) error {
	return m.createIfNotExists(l, ip)
}

func (m *MockRepository) GetByCode(_ context.Context, code string) (*link.Link, error) {
	return nil, nil
}

func TestService_CreateLink(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		ip        string
		url       string
		setupRepo func(m *MockRepository)
		setupGen  func(m *MockCodeGenerator)
		wantErr   error
		wantCode  string
	}{
		{
			name: "success on first try",
			ip:   "127.0.0.1",
			url:  "https://example.com",
			setupRepo: func(m *MockRepository) {
				m.countByIPFunc = func(ip string) (int, error) { return 0, nil }
				m.createIfNotExists = func(l *link.Link, ip string) error { return nil }
			},
			setupGen: func(m *MockCodeGenerator) {
				m.codes = []string{"CODE01"}
			},
			wantErr:  nil,
			wantCode: "CODE01",
		},
		{
			name: "success after one collision",
			ip:   "127.0.0.1",
			url:  "https://example.com",
			setupRepo: func(m *MockRepository) {
				m.countByIPFunc = func(ip string) (int, error) { return 0, nil }
				calls := 0
				m.createIfNotExists = func(l *link.Link, ip string) error {
					if calls == 0 {
						calls++
						return repository.ErrConflict
					}
					return nil
				}
			},
			setupGen: func(m *MockCodeGenerator) {
				m.codes = []string{"COLLID", "SUCCESS"}
			},
			wantErr:  nil,
			wantCode: "SUCCESS",
		},
		{
			name: "too many collisions",
			ip:   "127.0.0.1",
			url:  "https://example.com",
			setupRepo: func(m *MockRepository) {
				m.countByIPFunc = func(ip string) (int, error) { return 0, nil }
				m.createIfNotExists = func(l *link.Link, ip string) error { return repository.ErrConflict }
			},
			setupGen: func(m *MockCodeGenerator) {
				m.codes = []string{"C1", "C2", "C3", "C4", "C5"}
			},
			wantErr: link.ErrTooManyCollisions,
		},
		{
			name: "ip limit exceeded",
			ip:   "127.0.0.1",
			url:  "https://example.com",
			setupRepo: func(m *MockRepository) {
				m.countByIPFunc = func(ip string) (int, error) { return 10, nil }
			},
			setupGen: func(m *MockCodeGenerator) {},
			wantErr:  link.ErrTooManyActiveURLs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{}
			gen := &MockCodeGenerator{}
			tt.setupRepo(repo)
			tt.setupGen(gen)

			s := service.NewService(gen, repo)
			l, err := s.CreateLink(ctx, tt.ip, tt.url)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateLink() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr == nil {
				if l.Code != tt.wantCode {
					t.Errorf("CreateLink() code = %v, wantCode %v", l.Code, tt.wantCode)
				}
			}
		})
	}
}

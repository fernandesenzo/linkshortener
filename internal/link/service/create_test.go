package service_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
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
	countAndIncFunc func(ip string) (int, error)
	createFunc      func(l *link.Link, ip string) error
	decrementFunc   func(ip string) error

	DecrementCalls int
}

func (m *MockRepository) CountByIPAndIncrement(_ context.Context, ip string) (int, error) {
	return m.countAndIncFunc(ip)
}

func (m *MockRepository) CreateIfNotExists(_ context.Context, l *link.Link, ip string) error {
	return m.createFunc(l, ip)
}

func (m *MockRepository) DecrementIPCounter(_ context.Context, ip string) error {
	m.DecrementCalls++
	if m.decrementFunc != nil {
		return m.decrementFunc(ip)
	}
	return nil
}

func (m *MockRepository) GetByCode(_ context.Context, code string) (*link.Link, error) {
	return nil, nil
}

func TestService_CreateLink(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		ip             string
		url            string
		setupRepo      func(m *MockRepository)
		setupGen       func(m *MockCodeGenerator)
		wantErr        error
		wantCode       string
		wantDecrements int
	}{
		{
			name: "success on first try",
			ip:   "127.0.0.1",
			url:  "https://example.com",
			setupRepo: func(m *MockRepository) {
				m.countAndIncFunc = func(ip string) (int, error) { return 1, nil }
				m.createFunc = func(l *link.Link, ip string) error { return nil }
			},
			setupGen: func(m *MockCodeGenerator) {
				m.codes = []string{"CODE01"}
			},
			wantErr:        nil,
			wantCode:       "CODE01",
			wantDecrements: 0,
		},
		{
			name: "success after one collision",
			ip:   "127.0.0.1",
			url:  "https://example.com",
			setupRepo: func(m *MockRepository) {
				m.countAndIncFunc = func(ip string) (int, error) { return 1, nil }
				calls := 0
				m.createFunc = func(l *link.Link, ip string) error {
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
			wantErr:        nil,
			wantCode:       "SUCCESS",
			wantDecrements: 0,
		},
		{
			name: "too many collisions calls decrement",
			ip:   "127.0.0.1",
			url:  "https://example.com",
			setupRepo: func(m *MockRepository) {
				m.countAndIncFunc = func(ip string) (int, error) { return 1, nil }
				m.createFunc = func(l *link.Link, ip string) error { return repository.ErrConflict }
			},
			setupGen: func(m *MockCodeGenerator) {
				m.codes = []string{"C1", "C2", "C3", "C4", "C5"}
			},
			wantErr:        link.ErrTooManyCollisions,
			wantDecrements: 1,
		},
		{
			name: "ip limit exceeded calls decrement",
			ip:   "127.0.0.1",
			url:  "https://example.com",
			setupRepo: func(m *MockRepository) {
				m.countAndIncFunc = func(ip string) (int, error) { return 11, nil }
			},
			setupGen:       func(m *MockCodeGenerator) {},
			wantErr:        link.ErrTooManyActiveURLs,
			wantDecrements: 1,
		},
		{
			name: "repo error calls decrement",
			ip:   "127.0.0.1",
			url:  "https://example.com",
			setupRepo: func(m *MockRepository) {
				m.countAndIncFunc = func(ip string) (int, error) { return 1, nil }
				m.createFunc = func(l *link.Link, ip string) error { return errors.New("db error") }
			},
			setupGen: func(m *MockCodeGenerator) {
				m.codes = []string{"CODE01"}
			},
			wantErr:        repository.ErrConflict, // Dummy, will check err != nil
			wantDecrements: 1,
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

			if tt.wantErr != nil {
				if tt.name == "repo error calls decrement" {
					if err == nil {
						t.Errorf("CreateLink() expected error, got nil")
					}
				} else if !errors.Is(err, tt.wantErr) {
					t.Fatalf("CreateLink() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Fatalf("CreateLink() unexpected error = %v", err)
			}

			if tt.wantDecrements != repo.DecrementCalls {
				t.Errorf("DecrementIPCounter called %d times, want %d", repo.DecrementCalls, tt.wantDecrements)
			}

			if tt.wantErr == nil && err == nil {
				if l.Code != tt.wantCode {
					t.Errorf("CreateLink() code = %v, wantCode %v", l.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestService_CreateLink_Concurrency(t *testing.T) {
	ctx := context.Background()
	initialCount := 9
	limit := 10
	numRequests := 50

	dbCount := int32(initialCount)
	var successes int32
	var failures int32

	repo := &MockRepository{
		countAndIncFunc: func(ip string) (int, error) {
			newVal := atomic.AddInt32(&dbCount, 1)
			return int(newVal), nil
		},
		createFunc: func(l *link.Link, ip string) error {
			return nil
		},
		decrementFunc: func(ip string) error {
			atomic.AddInt32(&dbCount, -1)
			return nil
		},
	}

	gen := &MockCodeGenerator{codes: make([]string, numRequests)}
	for i := 0; i < numRequests; i++ {
		gen.codes[i] = fmt.Sprintf("CODE%d", i)
	}

	s := service.NewService(gen, repo)

	var wg sync.WaitGroup
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.CreateLink(ctx, "127.0.0.1", "https://google.com")
			if err == nil {
				atomic.AddInt32(&successes, 1)
			} else if errors.Is(err, link.ErrTooManyActiveURLs) {
				atomic.AddInt32(&failures, 1)
			}
		}()
	}
	wg.Wait()

	expectedSuccesses := limit - initialCount
	if int(successes) != expectedSuccesses {
		t.Errorf("expected %d successes, got %d", expectedSuccesses, successes)
	}

	if int(dbCount) != limit {
		t.Errorf("final count should be %d, got %d", limit, dbCount)
	}

	t.Logf("successes: %d, fails: %d, finalcount: %d", successes, failures, dbCount)
}

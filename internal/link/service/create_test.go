package service_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fernandesenzo/linkshortener/internal/link"
	"github.com/fernandesenzo/linkshortener/internal/link/repository"
	"github.com/fernandesenzo/linkshortener/internal/link/service"
)

func TestService_CreateLink(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name               string
		ip                 string
		url                string
		setupRepo          func(m *MockRepository)
		setupGen           func(m *MockCodeGenerator)
		wantErr            error
		wantCode           string
		wantIncrements     int
		wantGetIPLockCalls int
		wantUnlockCalls    int
	}{
		{
			name: "success on first try",
			ip:   "127.0.0.1",
			url:  "https://example.com",
			setupRepo: func(m *MockRepository) {
				m.countByIPFunc = func(ctx context.Context, ip string) (int, error) { return 1, nil }
				m.createFunc = func(ctx context.Context, l *link.Link, ip string) error { return nil }
			},
			setupGen: func(m *MockCodeGenerator) {
				m.codes = []string{"CODE01"}
			},
			wantErr:            nil,
			wantCode:           "CODE01",
			wantIncrements:     1,
			wantGetIPLockCalls: 1,
			wantUnlockCalls:    1,
		},
		{
			name: "success after one collision",
			ip:   "127.0.0.1",
			url:  "https://example.com",
			setupRepo: func(m *MockRepository) {
				m.countByIPFunc = func(ctx context.Context, ip string) (int, error) { return 1, nil }
				calls := 0
				m.createFunc = func(ctx context.Context, l *link.Link, ip string) error {
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
			wantErr:            nil,
			wantCode:           "SUCCESS",
			wantIncrements:     1,
			wantGetIPLockCalls: 1,
			wantUnlockCalls:    1,
		},
		{
			name: "too many collisions",
			ip:   "127.0.0.1",
			url:  "https://example.com",
			setupRepo: func(m *MockRepository) {
				m.countByIPFunc = func(ctx context.Context, ip string) (int, error) { return 1, nil }
				m.createFunc = func(ctx context.Context, l *link.Link, ip string) error { return repository.ErrConflict }
			},
			setupGen: func(m *MockCodeGenerator) {
				m.codes = []string{"C1", "C2", "C3", "C4", "C5"}
			},
			wantErr:            link.ErrTooManyCollisions,
			wantIncrements:     0,
			wantGetIPLockCalls: 1,
			wantUnlockCalls:    1,
		},
		{
			name: "ip limit exceeded",
			ip:   "127.0.0.1",
			url:  "https://example.com",
			setupRepo: func(m *MockRepository) {
				m.countByIPFunc = func(ctx context.Context, ip string) (int, error) { return 11, nil }
			},
			setupGen:           func(m *MockCodeGenerator) {},
			wantErr:            link.ErrTooManyActiveURLs,
			wantIncrements:     0,
			wantGetIPLockCalls: 1,
			wantUnlockCalls:    1,
		},
		{
			name: "repo error during creation",
			ip:   "127.0.0.1",
			url:  "https://example.com",
			setupRepo: func(m *MockRepository) {
				m.countByIPFunc = func(ctx context.Context, ip string) (int, error) { return 1, nil }
				m.createFunc = func(ctx context.Context, l *link.Link, ip string) error { return errors.New("db error") }
			},
			setupGen: func(m *MockCodeGenerator) {
				m.codes = []string{"CODE01"}
			},
			wantErr:            repository.ErrConflict, // dummy
			wantIncrements:     0,
			wantGetIPLockCalls: 1,
			wantUnlockCalls:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{}
			gen := &MockCodeGenerator{}
			tt.setupRepo(repo)
			tt.setupGen(gen)

			s := service.New(gen, repo)
			l, err := s.CreateLink(ctx, tt.ip, tt.url)

			if tt.wantErr != nil {
				if tt.name == "repo error during creation" {
					if err == nil {
						t.Errorf("CreateLink() expected error, got nil")
					}
				} else if !errors.Is(err, tt.wantErr) {
					t.Fatalf("CreateLink() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Fatalf("CreateLink() unexpected error = %v", err)
			}

			if tt.wantIncrements != repo.IncrementIPCounterCalls {
				t.Errorf("IncrementIPCounter called %d times, want %d", repo.IncrementIPCounterCalls, tt.wantIncrements)
			}

			if tt.wantGetIPLockCalls != repo.GetIPLockCalls {
				t.Errorf("GetIPLock called %d times, want %d", repo.GetIPLockCalls, tt.wantGetIPLockCalls)
			}

			if tt.wantUnlockCalls != repo.UnlockCalls {
				t.Errorf("Unlock called %d times, want %d", repo.UnlockCalls, tt.wantUnlockCalls)
			}

			if tt.wantErr == nil && err == nil {
				if l.Code != tt.wantCode {
					t.Errorf("CreateLink() code = %v, wantCode %v", l.Code, tt.wantCode)
				}
				diff := time.Until(l.ExpiresAt) - link.DefaultTTL
				if diff < 0 {
					diff = -diff
				}
				if diff > 5*time.Second {
					t.Errorf("CreateLink() ExpiresAt = %v, want roughly %v (diff = %v)", l.ExpiresAt, time.Now().Add(link.DefaultTTL), diff)
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

	var mu sync.Mutex
	dbCount := initialCount
	var successes int32
	var failures int32

	repo := &MockRepository{
		getIPLockFunc: func(ctx context.Context, ip string) (func(), error) {
			mu.Lock()
			return func() {
				mu.Unlock()
			}, nil
		},
		countByIPFunc: func(ctx context.Context, ip string) (int, error) {
			return dbCount, nil
		},
		createFunc: func(ctx context.Context, l *link.Link, ip string) error {
			return nil
		},
		incrementIPCounterFunc: func(ctx context.Context, ip string) error {
			dbCount++
			return nil
		},
	}

	gen := &MockCodeGenerator{codes: make([]string, numRequests)}
	for i := 0; i < numRequests; i++ {
		gen.codes[i] = fmt.Sprintf("CODE%d", i)
	}

	s := service.New(gen, repo)

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

	if dbCount != limit {
		t.Errorf("final count should be %d, got %d", limit, dbCount)
	}

	t.Logf("successes: %d, fails: %d, finalcount: %d", successes, failures, dbCount)
}

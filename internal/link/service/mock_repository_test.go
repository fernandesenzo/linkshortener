package service_test

import (
	"context"
	"sync"

	"github.com/fernandesenzo/linkshortener/internal/link"
	"github.com/fernandesenzo/linkshortener/internal/link/repository"
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
	callsMu sync.Mutex

	getIPLockFunc func(ctx context.Context, ip string) (func(), error)
	createFunc    func(ctx context.Context, l *link.Link, ip string) error
	countByIPFunc func(ctx context.Context, ip string) (int, error)
	getByCodeFunc func(ctx context.Context, code string) (*link.Link, error)

	GetIPLockCalls int
	UnlockCalls    int
	CreateCalls    int
	CountByIPCalls int
	GetByCodeCalls int
}

func (m *MockRepository) incrementCalls(counter *int) {
	m.callsMu.Lock()
	defer m.callsMu.Unlock()
	(*counter)++
}

func (m *MockRepository) GetIPLock(ctx context.Context, ip string) (unlock func(), err error) {
	m.incrementCalls(&m.GetIPLockCalls)
	if m.getIPLockFunc != nil {
		return m.getIPLockFunc(ctx, ip)
	}
	return func() {
		m.incrementCalls(&m.UnlockCalls)
	}, nil
}

func (m *MockRepository) Create(ctx context.Context, l *link.Link, ip string) error {
	m.incrementCalls(&m.CreateCalls)
	if m.createFunc != nil {
		return m.createFunc(ctx, l, ip)
	}
	return nil
}

func (m *MockRepository) CountByIP(ctx context.Context, ip string) (int, error) {
	m.incrementCalls(&m.CountByIPCalls)
	if m.countByIPFunc != nil {
		return m.countByIPFunc(ctx, ip)
	}
	return 0, nil
}

func (m *MockRepository) GetByCode(ctx context.Context, code string) (*link.Link, error) {
	m.incrementCalls(&m.GetByCodeCalls)
	if m.getByCodeFunc != nil {
		return m.getByCodeFunc(ctx, code)
	}
	return nil, repository.ErrNotFound
}

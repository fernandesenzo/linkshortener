package service_test

import (
	"context"

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
	getIPLockFunc          func(ctx context.Context, ip string) (func(), error)
	createFunc             func(ctx context.Context, l *link.Link, ip string) error
	countByIPFunc          func(ctx context.Context, ip string) (int, error)
	getByCodeFunc          func(ctx context.Context, code string) (*link.Link, error)
	incrementIPCounterFunc func(ctx context.Context, ip string) error

	GetIPLockCalls          int
	UnlockCalls             int
	CreateCalls             int
	CountByIPCalls          int
	GetByCodeCalls          int
	IncrementIPCounterCalls int
}

func (m *MockRepository) GetIPLock(ctx context.Context, ip string) (unlock func(), err error) {
	m.GetIPLockCalls++
	if m.getIPLockFunc != nil {
		return m.getIPLockFunc(ctx, ip)
	}
	return func() {
		m.UnlockCalls++
	}, nil
}

func (m *MockRepository) Create(ctx context.Context, l *link.Link, ip string) error {
	m.CreateCalls++
	if m.createFunc != nil {
		return m.createFunc(ctx, l, ip)
	}
	return nil
}

func (m *MockRepository) CountByIP(ctx context.Context, ip string) (int, error) {
	m.CountByIPCalls++
	if m.countByIPFunc != nil {
		return m.countByIPFunc(ctx, ip)
	}
	return 0, nil
}

func (m *MockRepository) GetByCode(ctx context.Context, code string) (*link.Link, error) {
	m.GetByCodeCalls++
	if m.getByCodeFunc != nil {
		return m.getByCodeFunc(ctx, code)
	}
	return nil, repository.ErrNotFound
}

func (m *MockRepository) IncrementIPCounter(ctx context.Context, ip string) error {
	m.IncrementIPCounterCalls++
	if m.incrementIPCounterFunc != nil {
		return m.incrementIPCounterFunc(ctx, ip)
	}
	return nil
}

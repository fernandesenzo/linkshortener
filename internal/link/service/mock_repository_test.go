package service_test

import (
	"context"

	"github.com/fernandesenzo/linkshortener/internal/link"
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
	getByCodeFunc   func(code string) (*link.Link, error)

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
	if m.getByCodeFunc != nil {
		return m.getByCodeFunc(code)
	}
	return nil, nil
}

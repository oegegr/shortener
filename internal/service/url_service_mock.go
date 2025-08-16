package service

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockURLService struct {
	mock.Mock
}

func (m *MockURLService) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	args := m.Called(originalURL)
	return args.String(0), args.Error(1)
}

func (m *MockURLService) GetShortURLBatch(ctx context.Context, urls []string) ([]string, error) {
	args := m.Called(urls)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockURLService) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	args := m.Called(shortURL)
	return args.String(0), args.Error(1)
}

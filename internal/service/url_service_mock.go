package service

import (
	"context"

	"github.com/oegegr/shortener/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockURLService struct {
	mock.Mock
}

func (m *MockURLService) GetShortURL(ctx context.Context, originalURL string, userID string) (string, error) {
	args := m.Called(ctx, originalURL, userID)
	return args.String(0), args.Error(1)
}

func (m *MockURLService) GetShortURLBatch(ctx context.Context, urls []string, userID string) ([]string, error) {
	args := m.Called(ctx, urls, userID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockURLService) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	args := m.Called(ctx, shortURL)
	return args.String(0), args.Error(1)
}

func (m *MockURLService) GetUserURL(ctx context.Context, userID string) ([]model.UserURL, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.UserURL), args.Error(1)
}

func (m *MockURLService) DeleteUserURL(ctx context.Context, userID string, shortIDs []string) error {
	args := m.Called(ctx, userID, shortIDs)
	return args.Error(1)
}
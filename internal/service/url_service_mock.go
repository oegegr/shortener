// Package service содержит мок-реализацию сервиса сокращения URL-адресов для тестирования.
package service

import (
	"context"

	"github.com/oegegr/shortener/internal/model"
	"github.com/stretchr/testify/mock"
)

// MockURLService представляет мок-реализацию сервиса сокращения URL-адресов.
type MockURLService struct {
	mock.Mock
}

// GetShortURL возвращает сокращенный URL-адрес для заданного URL-адреса и идентификатора пользователя (мок-реализация).
func (m *MockURLService) GetShortURL(ctx context.Context, originalURL string, userID string) (string, error) {
	args := m.Called(ctx, originalURL, userID)
	return args.String(0), args.Error(1)
}

// GetShortURLBatch возвращает список сокращенных URL-адресов для заданного списка URL-адресов и идентификатора пользователя (мок-реализация).
func (m *MockURLService) GetShortURLBatch(ctx context.Context, urls []string, userID string) ([]string, error) {
	args := m.Called(ctx, urls, userID)
	return args.Get(0).([]string), args.Error(1)
}

// GetOriginalURL возвращает оригинальный URL-адрес для заданного сокращенного URL-адреса (мок-реализация).
func (m *MockURLService) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	args := m.Called(ctx, shortURL)
	return args.String(0), args.Error(1)
}

// GetUserURL возвращает список URL-адресов для заданного идентификатора пользователя (мок-реализация).
func (m *MockURLService) GetUserURL(ctx context.Context, userID string) ([]model.UserURL, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.UserURL), args.Error(1)
}

// DeleteUserURL удаляет URL-адреса для заданного идентификатора пользователя и списка сокращенных URL-адресов (мок-реализация).
func (m *MockURLService) DeleteUserURL(ctx context.Context, userID string, shortIDs []string) error {
	args := m.Called(ctx, userID, shortIDs)
	return args.Error(1)
}

// GetStat возвращает статистику использования сервиса.
func (m *MockURLService) GetStats(ctx context.Context) (*model.Stats, error) {
	args := m.Called(ctx)
	stats := args.Get(0).(model.Stats)
	return &stats, args.Error(1)
}

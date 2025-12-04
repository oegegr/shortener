// Package repository содержит мок-реализацию репозитория URL-адресов для тестирования.
package repository

import (
	"context"

	"github.com/oegegr/shortener/internal/model"
	"github.com/stretchr/testify/mock"
)

// MockURLRepository представляет мок-реализацию репозитория URL-адресов.
type MockURLRepository struct {
	mock.Mock
}

// Ping проверяет подключение к репозиторию (мок-реализация).
func (m *MockURLRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// CreateURL создает новые URL-адреса в репозитории (мок-реализация).
func (m *MockURLRepository) CreateURL(ctx context.Context, urls []model.URLItem) error {
	args := m.Called(ctx, urls)
	return args.Error(0)
}

// DeleteURL удаляет URL-адреса из репозитория (мок-реализация).
func (m *MockURLRepository) DeleteURL(ctx context.Context, ids []string) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

// FindURLByID находит URL-адрес в репозитории по идентификатору URL-адреса (мок-реализация).
func (m *MockURLRepository) FindURLByID(ctx context.Context, id string) (*model.URLItem, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.URLItem), args.Error(1)
}

// FindURLByURL находит URL-адрес в репозитории по URL-адресу (мок-реализация).
func (m *MockURLRepository) FindURLByURL(ctx context.Context, url string) (*model.URLItem, error) {
	args := m.Called(ctx, url)
	return args.Get(0).(*model.URLItem), args.Error(1)
}

// FindURLByUser находит URL-адреса в репозитории по идентификатору пользователя (мок-реализация).
func (m *MockURLRepository) FindURLByUser(ctx context.Context, id string) ([]model.URLItem, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]model.URLItem), args.Error(1)
}

// Exists проверяет, существует ли URL-адрес в репозитории (мок-реализация).
func (m *MockURLRepository) Exists(ctx context.Context, id string) bool {
	args := m.Called(ctx, id)
	return args.Bool(0)
}

// GetUserCount возращает кол-во пользователей
func (m *MockURLRepository) GetUserCount(ctx context.Context) (*int, error) {
	args := m.Called(ctx)
	count := args.Int(0)
	return &count, args.Error(1)

}

// GetURLCount возращает кол-во сокращенных URL
func (m *MockURLRepository) GetURLCount(ctx context.Context) (*int, error) {
	args := m.Called(ctx)
	count := args.Int(0)
	return &count, args.Error(1)

}

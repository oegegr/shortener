package repository

import (
	"context"

	"github.com/oegegr/shortener/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockURLRepository struct {
	mock.Mock
}

func (m *MockURLRepository) Ping(ctx context.Context) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockURLRepository) CreateURL(ctx context.Context, urls []model.URLItem) error {
	args := m.Called(urls)
	return args.Error(0)
}

func (m *MockURLRepository) FindURLByID(ctx context.Context, id string) (*model.URLItem, error) {
	args := m.Called(id)
	return args.Get(0).(*model.URLItem), args.Error(1)
}

func (m *MockURLRepository) FindURLByURL(ctx context.Context, url string) (*model.URLItem, error) {
	args := m.Called(url)
	return args.Get(0).(*model.URLItem), args.Error(1)
}

func (m *MockURLRepository) Exists(ctx context.Context, id string) bool {
	args := m.Called(id)
	return args.Bool(0)
}

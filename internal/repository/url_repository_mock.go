package repository

import (
	"github.com/oegegr/shortener/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockURLRepository struct {
	mock.Mock
}

func (m *MockURLRepository) CreateURL(url model.URLItem) error {
	args := m.Called(url)
	return args.Error(0)
}

func (m *MockURLRepository) FindURLByID(id string) (*model.URLItem, error) {
	args := m.Called(id)
	return args.Get(0).(*model.URLItem), args.Error(1)
}

func (m *MockURLRepository) Exists(id string) bool {
	args := m.Called(id)
	return args.Bool(0)
}

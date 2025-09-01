package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/oegegr/shortener/internal/model"
	"github.com/oegegr/shortener/internal/repository"
	"github.com/oegegr/shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

type MockShortCodeProvider struct {
	mock.Mock
}

func (m *MockShortCodeProvider) Get(length int) string {
	args := m.Called(length)
	return args.String(0)
}

const (
	user string = "test"
)

func TestShortenURLService_GetShortURL_Success(t *testing.T) {
	repoMock := new(repository.MockURLRepository)
	provider := new(MockShortCodeProvider)
	ctx := context.Background()
	logger := zaptest.NewLogger(t).Sugar()
	svc := service.NewShortenerService(repoMock, "https://short.com", 6, provider, *logger)

	originalURL := "https://original.com/long/url"
	expectedShortCode := "abc123"

	repoMock.On("CreateURL", mock.AnythingOfType("[]model.URLItem")).Return(nil).Once()
	provider.On("Get", 6).Return(expectedShortCode)

	shortURL, err := svc.GetShortURL(ctx, originalURL, user)

	assert.NoError(t, err)
	assert.Equal(t, "https://short.com/"+expectedShortCode, shortURL)
	repoMock.AssertExpectations(t)
}

func TestShortenURLService_GetShortURL_CollisionRecovery(t *testing.T) {
	repoMock := new(repository.MockURLRepository)
	provider := new(MockShortCodeProvider)
	ctx := context.Background()
	logger := zaptest.NewLogger(t).Sugar()
	svc := service.NewShortenerService(repoMock, "https://short.com", 6, provider, *logger)

	originalURL := "https://original.com/long/url"

	repoMock.On("CreateURL", mock.Anything).Return(repository.ErrRepoShortIDAlreadyExists).Twice()
	repoMock.On("CreateURL", mock.Anything).Return(nil).Once()
	provider.On("Get", 6).Return("any")

	shortURL, err := svc.GetShortURL(ctx, originalURL, user)

	assert.NoError(t, err)
	assert.Contains(t, shortURL, "https://short.com/")
	repoMock.AssertExpectations(t)
	repoMock.AssertNumberOfCalls(t, "CreateURL", 3)
}

func TestShortenURLService_GetShortURL_MaxCollisions(t *testing.T) {
	repoMock := new(repository.MockURLRepository)
	provider := new(MockShortCodeProvider)
	ctx := context.Background()
	logger := zaptest.NewLogger(t).Sugar()
	svc := service.NewShortenerService(repoMock, "https://short.com", 6, provider, *logger)

	originalURL := "https://original.com/long/url"

	repoMock.On("CreateURL", mock.Anything).Return(repository.ErrRepoShortIDAlreadyExists).Times(10)
	provider.On("Get", 6).Return("any")

	shortURL, err := svc.GetShortURL(ctx, originalURL, user)

	assert.Error(t, err)
	assert.Equal(t, repository.ErrRepoShortIDAlreadyExists, err)
	assert.Empty(t, shortURL)
	repoMock.AssertExpectations(t)
}

func TestShortenURLService_GetShortURL_RepositoryError(t *testing.T) {
	repoMock := new(repository.MockURLRepository)
	provider := new(MockShortCodeProvider)
	ctx := context.Background()
	logger := zaptest.NewLogger(t).Sugar()
	svc := service.NewShortenerService(repoMock, "https://short.com", 6, provider, *logger)

	originalURL := "https://original.com/long/url"
	testError := errors.New("database failure")

	repoMock.On("CreateURL", mock.Anything).Return(testError)
	provider.On("Get", 6).Return("any")

	shortURL, err := svc.GetShortURL(ctx, originalURL, user)

	assert.Error(t, err)
	assert.Equal(t, testError, err)
	assert.Empty(t, shortURL)
	repoMock.AssertExpectations(t)
}

func TestShortenURLService_GetOriginalURL_Success(t *testing.T) {
	repoMock := new(repository.MockURLRepository)
	provider := new(MockShortCodeProvider)
	ctx := context.Background()
	logger := zaptest.NewLogger(t).Sugar()
	svc := service.NewShortenerService(repoMock, "https://short.com", 6, provider, *logger)

	shortCode := "abc123"
	expectedURL := "https://original.com/long/url"
	urlItem := &model.URLItem{ShortID: shortCode, URL: expectedURL}

	repoMock.On("FindURLByID", shortCode).Return(urlItem, nil).Once()

	originalURL, err := svc.GetOriginalURL(ctx, shortCode)

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, originalURL)
	repoMock.AssertExpectations(t)
}

func TestShortenURLService_GetOriginalURL_NotFound(t *testing.T) {
	repoMock := new(repository.MockURLRepository)
	provider := new(MockShortCodeProvider)
	ctx := context.Background()
	logger := zaptest.NewLogger(t).Sugar()
	svc := service.NewShortenerService(repoMock, "https://short.com", 6, provider, *logger)

	shortCode := "invalid123"

	repoMock.On("FindURLByID", shortCode).Return(&model.URLItem{}, repository.ErrRepoNotFound).Once()

	originalURL, err := svc.GetOriginalURL(ctx, shortCode)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, repository.ErrRepoNotFound))
	assert.Empty(t, originalURL)
	repoMock.AssertExpectations(t)
}

func TestShortenURLService_GetOriginalURL_RepositoryError(t *testing.T) {
	repoMock := new(repository.MockURLRepository)
	provider := new(MockShortCodeProvider)
	ctx := context.Background()
	logger := zaptest.NewLogger(t).Sugar()
	svc := service.NewShortenerService(repoMock, "https://short.com", 6, provider, *logger)

	shortCode := "abc123"
	testError := errors.New("database error")

	repoMock.On("FindURLByID", shortCode).Return(&model.URLItem{}, testError)

	originalURL, err := svc.GetOriginalURL(ctx, shortCode)

	assert.Error(t, err)
	assert.Equal(t, testError, err)
	assert.Empty(t, originalURL)
	repoMock.AssertExpectations(t)
}

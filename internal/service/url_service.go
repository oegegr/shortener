// Package service содержит реализацию сервиса для работы с URL-адресами.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	app_error "github.com/oegegr/shortener/internal/error"
	"github.com/oegegr/shortener/internal/model"
	"github.com/oegegr/shortener/internal/repository"

	"github.com/avast/retry-go"
	"go.uber.org/zap"
)

// maxCollisionAttempts представляет максимальное количество попыток для разрешения коллизий.
const maxCollisionAttempts = 10

// retryCollisionTimeout представляет время ожидания между попытками для разрешения коллизий.
const retryCollisionTimeout = 100 * time.Millisecond

// URLShortener представляет интерфейс для сервиса сокращения URL-адресов.
type URLShortener interface {
	// GetShortURL возвращает сокращенный URL-адрес для заданного URL-адреса и идентификатора пользователя.
	GetShortURL(ctx context.Context, url string, userID string) (string, error)
	// GetShortURLBatch возвращает список сокращенных URL-адресов для заданного списка URL-адресов и идентификатора пользователя.
	GetShortURLBatch(ctx context.Context, urls []string, userID string) ([]string, error)
	// GetOriginalURL возвращает оригинальный URL-адрес для заданного сокращенного URL-адреса.
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
	// GetUserURL возвращает список URL-адресов для заданного идентификатора пользователя.
	GetUserURL(ctx context.Context, userID string) ([]model.UserURL, error)
	// DeleteUserURL удаляет URL-адреса для заданного идентификатора пользователя и списка сокращенных URL-адресов.
	DeleteUserURL(ctx context.Context, userID string, shortIDs []string) error
}

// URLDeletionStrategy представляет интерфейс для стратегии удаления URL-адресов.
type URLDeletionStrategy interface {
	// DeleteURL удаляет URL-адреса для заданного контекста и списка сокращенных URL-адресов.
	DeleteURL(ctx context.Context, shortIDs []string) error
}

// ShortenURLService представляет реализацию сервиса сокращения URL-адресов.
type ShortenURLService struct {
	// urlRepository представляет репозиторий URL-адресов.
	urlRepository repository.URLRepository
	// shortURLDomain представляет домен для сокращенных URL-адресов.
	shortURLDomain string
	// shortURLLength представляет длину сокращенного URL-адреса.
	shortURLLength int
	// shortCodeProvider представляет провайдер сокращенных кодов.
	shortCodeProvider ShortCodeProvider
	// urlDelStrategy представляет стратегию удаления URL-адресов.
	urlDelStrategy URLDeletionStrategy
	// logger представляет логгер для записи сообщений.
	logger zap.SugaredLogger
}

// NewShortenerService возвращает новый экземпляр ShortenURLService.
// Эта функция принимает репозиторий URL-адресов, домен для сокращенных URL-адресов, длину сокращенного URL-адреса, провайдер сокращенных кодов, стратегию удаления URL-адресов и логгер.
func NewShortenerService(
	repository repository.URLRepository,
	domain string,
	urlLength int,
	codeProvider ShortCodeProvider,
	urlDelStrategy URLDeletionStrategy,
	logger zap.SugaredLogger,
) *ShortenURLService {
	return &ShortenURLService{
		urlRepository:     repository,
		shortURLDomain:    domain,
		shortURLLength:    urlLength,
		shortCodeProvider: codeProvider,
		urlDelStrategy:    urlDelStrategy,
		logger:            logger,
	}
}

// GetShortURL возвращает сокращенный URL-адрес для заданного URL-адреса и идентификатора пользователя.
func (s *ShortenURLService) GetShortURL(ctx context.Context, url string, userID string) (string, error) {
	items, err := s.tryGetURLItem(ctx, []string{url}, userID)

	if err != nil {
		if errors.Is(err, repository.ErrRepoURLAlreadyExists) {
			return s.resolveURLConflict(ctx, url, err)
		}
		return "", err
	}

	return s.buildShortURL(items[0]), nil
}

// DeleteUserURL удаляет URL-адреса для заданного идентификатора пользователя и списка сокращенных URL-адресов.
func (s *ShortenURLService) DeleteUserURL(ctx context.Context, userID string, shortIDs []string) error {
	err := s.urlDelStrategy.DeleteURL(ctx, shortIDs)
	if err != nil {
		return err
	}
	return nil
}

// GetUserURL возвращает список URL-адресов для заданного идентификатора пользователя.
func (s *ShortenURLService) GetUserURL(ctx context.Context, userID string) ([]model.UserURL, error) {
	items, err := s.urlRepository.FindURLByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	urls := []model.UserURL{}
	for _, item := range items {
		userURL := model.UserURL{
			URL:      item.URL,
			ShortURL: s.buildShortURL(item),
		}
		urls = append(urls, userURL)
	}
	return urls, nil
}

// resolveURLConflict разрешает конфликт URL-адресов, возвращая сокращенный URL-адрес для заданного URL-адреса и ошибки.
func (s *ShortenURLService) resolveURLConflict(ctx context.Context, url string, urlConflict error) (string, error) {
	item, err := s.urlRepository.FindURLByURL(ctx, url)
	if err != nil {
		return "", err
	}
	return s.buildShortURL(*item), urlConflict
}

// GetShortURLBatch возвращает список сокращенных URL-адресов для заданного списка URL-адресов и идентификатора пользователя.
func (s *ShortenURLService) GetShortURLBatch(ctx context.Context, urls []string, userID string) ([]string, error) {
	shorts := []string{}
	items, err := s.tryGetURLItem(ctx, urls, userID)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		shorts = append(shorts, s.buildShortURL(item))
	}
	return shorts, nil
}

// GetOriginalURL возвращает оригинальный URL-адрес для заданного сокращенного URL-адреса.
func (s *ShortenURLService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	urlItem, err := s.urlRepository.FindURLByID(ctx, shortCode)

	if err != nil {
		return "", err
	}

	if urlItem.IsDeleted {
		return "", app_error.ErrServiceURLGone
	}

	return urlItem.URL, nil
}

// getURLItem возвращает список элементов URL-адресов для заданного списка URL-адресов и идентификатора пользователя.
func (s *ShortenURLService) getURLItem(ctx context.Context, originalURL []string, userID string) ([]model.URLItem, error) {
	items := []model.URLItem{}
	for _, url := range originalURL {
		shortCode := s.shortCodeProvider.Get(s.shortURLLength)
		item := model.NewURLItem(url, shortCode, userID, false)
		items = append(items, *item)
	}
	err := s.urlRepository.CreateURL(ctx, items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// tryGetURLItem возвращает список элементов URL-адресов для заданного списка URL-адресов и идентификатора пользователя с повторными попытками в случае коллизий.
func (s *ShortenURLService) tryGetURLItem(ctx context.Context, originalURL []string, userID string) ([]model.URLItem, error) {
	var items []model.URLItem
	err := retry.Do(
		func() error {
			var err error
			items, err = s.getURLItem(ctx, originalURL, userID)
			return err
		},
		retry.RetryIf(
			func(err error) bool {
				return errors.Is(err, repository.ErrRepoShortIDAlreadyExists)
			},
		),
		retry.LastErrorOnly(true),
		retry.Attempts(maxCollisionAttempts),
		retry.MaxDelay(retryCollisionTimeout),
		retry.Context(ctx),
		retry.OnRetry(func(n uint, err error) { s.logger.Debugln("Retry error: ", err.Error()) }),
	)

	if err != nil {
		return nil, err
	}
	return items, nil
}

// buildShortURL возвращает сокращенный URL-адрес для заданного элемента URL-адреса.
func (s *ShortenURLService) buildShortURL(item model.URLItem) string {
	return fmt.Sprintf("%s/%s", s.shortURLDomain, item.ShortID)
}

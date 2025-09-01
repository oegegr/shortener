package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/oegegr/shortener/internal/model"
	"github.com/oegegr/shortener/internal/repository"

	"github.com/avast/retry-go"
	"go.uber.org/zap"
)

const (
	maxCollisionAttempts  = 10
	retryCollisionTimeout = 100 * time.Millisecond
)

type URLShortener interface {
	GetShortURL(ctx context.Context, url string, userID string) (string, error)
	GetShortURLBatch(ctx context.Context, urls []string, userID string) ([]string, error)
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
	GetUserURL(ctx context.Context, userID string) ([]model.UserURL, error)
}

type ShortenURLService struct {
	urlRepository     repository.URLRepository
	shortURLDomain    string
	shortURLLength    int
	shortCodeProvider ShortCodeProvider
	logger            zap.SugaredLogger
}

func NewShortenerService(
	repository repository.URLRepository,
	domain string,
	urlLength int,
	codeProvider ShortCodeProvider,
	logger zap.SugaredLogger) *ShortenURLService {
	return &ShortenURLService{
		urlRepository:     repository,
		shortURLDomain:    domain,
		shortURLLength:    urlLength,
		shortCodeProvider: codeProvider,
		logger:            logger,
	}
}

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

func (s *ShortenURLService) GetUserURL(ctx context.Context, userID string) ([]model.UserURL, error) {
	items, err := s.urlRepository.FindURLByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	urls := []model.UserURL{}
	for _, item := range items {
		userURL := model.UserURL{
			URL: item.URL,
			ShortURL: s.buildShortURL(item),
		}
		urls = append(urls, userURL)
	}
	return urls, nil
}

func (s *ShortenURLService) resolveURLConflict(ctx context.Context, url string, urlConflict error) (string, error) {
	item, err := s.urlRepository.FindURLByURL(ctx, url)
	if err != nil {
		return "", err
	}
	return s.buildShortURL(*item), urlConflict 
}

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

func (s *ShortenURLService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	urlItem, err := s.urlRepository.FindURLByID(ctx, shortCode)
	if err != nil {
		return "", err
	}

	return urlItem.URL, nil
}

func (s *ShortenURLService) getURLItem(ctx context.Context, originalURL []string, userID string) ([]model.URLItem, error) {
	items := []model.URLItem{}
	for _, url := range originalURL {
		shortCode := s.shortCodeProvider.Get(s.shortURLLength)
		item := model.NewURLItem(url, shortCode, userID)
		items = append(items, *item)
	}
	err := s.urlRepository.CreateURL(ctx, items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

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

func (s *ShortenURLService) buildShortURL(item model.URLItem) string {
	return fmt.Sprintf("%s/%s", s.shortURLDomain, item.ShortID)
}

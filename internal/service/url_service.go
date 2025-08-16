package service

import (
	"context"
	"fmt"
	"time"

	app_error "github.com/oegegr/shortener/internal/error"
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
	GetShortURL(originalURL string) (string, error)
	GetOriginalURL(shortURL string) (string, error)
}

type ShortenURLService struct {
	urlRepository     repository.URLRepository
	shortURLDomain    string
	shortURLLength    int
	shortCodeProvider ShortCodeProvider
	ctx               context.Context
	logger            zap.SugaredLogger
}

func NewShortenerService(
	repository repository.URLRepository,
	domain string,
	urlLength int,
	codeProvider ShortCodeProvider,
	ctx context.Context,
	logger zap.SugaredLogger) *ShortenURLService {
	return &ShortenURLService{
		urlRepository:     repository,
		shortURLDomain:    domain,
		shortURLLength:    urlLength,
		shortCodeProvider: codeProvider,
		ctx:               ctx,
		logger:            logger,
	}
}

func (s *ShortenURLService) GetShortURL(originalURL string) (string, error) {
	urlItem, err := s.tryGetURLItem(originalURL)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", s.shortURLDomain, urlItem.ShortID), nil
}

func (s *ShortenURLService) GetOriginalURL(shortCode string) (string, error) {
	urlItem, err := s.urlRepository.FindURLByID(shortCode)
	if err != nil {
		return "", err
	}

	return urlItem.URL, nil
}

func (s *ShortenURLService) getURLItem(originalURL string) (*model.URLItem, error) {
	shortCode := s.shortCodeProvider.Get(s.shortURLLength)
	urlItem := model.NewURLItem(originalURL, shortCode)
	err := s.urlRepository.CreateURL(*urlItem)
	if err != nil {
		return nil, err
	}
	return urlItem, nil
}

func (s *ShortenURLService) tryGetURLItem(originalURL string) (*model.URLItem, error) {
	var urlItem *model.URLItem
	err := retry.Do(
		func() error {
			var err error
			urlItem, err = s.getURLItem(originalURL)
			return err
		},
		retry.Attempts(maxCollisionAttempts),
		retry.MaxDelay(retryCollisionTimeout),
		retry.Context(s.ctx),
		retry.OnRetry(func(n uint, err error) { s.logger.Debugln("Retry error: ", err.Error()) }),
	)

	if err != nil {
		return nil, app_error.ErrServiceFailedToGetShortURL
	}
	return urlItem, nil
}

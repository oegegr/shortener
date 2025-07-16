package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/oegegr/shortener/internal/model"
	"github.com/oegegr/shortener/internal/repository"
)

var (
	maxCollisionAttempts          = 10
	retryCollisionTimeout, _      = time.ParseDuration("0.1s")
	ErrServiceFailedToGetShortURL = errors.New("failed to get short url")
)

type URLShortner interface {
	GetShortURL(originalURL string) (string, error)
	GetOriginalURL(shortURL string) (string, error)
}

type ShortenURLService struct {
	urlRepository  repository.URLRepository
	shortURLDomain string
	shortURLLength int
}

func NewShortnerService(repository repository.URLRepository, domain string, urlLength int) *ShortenURLService {
	return &ShortenURLService{urlRepository: repository, shortURLDomain: domain, shortURLLength: urlLength}
}

func (s *ShortenURLService) GetShortURL(originalURL string) (string, error) {
	urlItem, err := s.tryGetURLItem(originalURL)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", s.shortURLDomain, urlItem.ID), nil
}

func (s *ShortenURLService) GetOriginalURL(shortCode string) (string, error) {
	urlItem, err := s.urlRepository.FindURLById(shortCode)
	if err != nil {
		return "", err
	}

	return urlItem.URL, nil
}

func (s *ShortenURLService) getURLItem(originalURL string) (*model.URLItem, error) {
	shortCode := GenerateShortCode(s.shortURLLength)
	urlItem := model.NewURLItem(originalURL, shortCode)
	err := s.urlRepository.CreateURL(*urlItem)
	if err != nil {
		return nil, err
	}
	return urlItem, nil
}

func (s *ShortenURLService) tryGetURLItem(originalURL string) (*model.URLItem, error) {
	var err error
	for range maxCollisionAttempts {
		urlItem, err := s.getURLItem(originalURL)
		if err == nil {
			return urlItem, nil
		}
		if !errors.Is(err, repository.ErrRepoAlreadyExists) {
			return nil, err
		}
		time.Sleep(retryCollisionTimeout)
	}
	return nil, err
}

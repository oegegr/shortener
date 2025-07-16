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
	ErrServiceFailedToGetShortUrl = errors.New("failed to get short url")
)

type URLShortner interface {
	GetShortURL(originalUrl string) (string, error)
	GetOriginalURL(shortUrl string) (string, error)
}

type ShortenURLService struct {
	urlRepository  repository.URLRepository
	shortUrlDomain string
	shortUrlLength int
}

func NewShortnerService(repository repository.URLRepository, domain string, urlLength int) *ShortenURLService {
	return &ShortenURLService{urlRepository: repository, shortUrlDomain: domain, shortUrlLength: urlLength}
}

func (s *ShortenURLService) GetShortURL(originalUrl string) (string, error) {
	urlItem, err := s.tryGetURLItem(originalUrl)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", s.shortUrlDomain, urlItem.ID), nil
}

func (s *ShortenURLService) GetOriginalURL(shortCode string) (string, error) {
	urlItem, err := s.urlRepository.FindURLById(shortCode)
	if err != nil {
		return "", err
	}

	return urlItem.URL, nil
}

func (s *ShortenURLService) getURLItem(originalUrl string) (*model.UrlItem, error) {
	shortCode := GenerateShortCode(s.shortUrlLength)
	urlItem := model.NewURLItem(originalUrl, shortCode)
	err := s.urlRepository.CreateURL(*urlItem)
	if err != nil {
		return nil, err
	}
	return urlItem, nil
}

func (s *ShortenURLService) tryGetURLItem(originalUrl string) (*model.UrlItem, error) {
	var err error
	for range maxCollisionAttempts {
		urlItem, err := s.getURLItem(originalUrl)
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

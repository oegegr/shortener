package service

import (
	"fmt"
	"time"
	"errors"

	"github.com/oegegr/shortener/internal/model"
	"github.com/oegegr/shortener/internal/repository"
)

var (
	maxCollisionAttempts = 10
	retryCollisionTimeout, _ = time.ParseDuration("0.1s")
	ErrServiceFailedToGetShortUrl = errors.New("failed to get short url")
)

type UrlShortner interface {
	GetShortUrl(originalUrl string) (string, error)
	GetOriginalUrl(shortUrl string) (string, error)
}

type ShortenUrlService struct {
	urlRepository  repository.UrlRepository
	shortUrlDomain string
	shortUrlLength int
}

func NewShortnerService(repository repository.UrlRepository, domain string, urlLength int) *ShortenUrlService {
	return &ShortenUrlService{urlRepository: repository, shortUrlDomain: domain, shortUrlLength: urlLength}
}

func (s *ShortenUrlService) GetShortUrl(originalUrl string) (string, error) {
	urlItem, err := s.tryGetUrlItem(originalUrl)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", s.shortUrlDomain, urlItem.Id), nil
}

func (s *ShortenUrlService) GetOriginalUrl(shortCode string) (string, error) {
	urlItem, err := s.urlRepository.FindUrlById(shortCode)
	if err != nil {
		return "", err
	}

	return urlItem.Url, nil
}

func (s *ShortenUrlService) getUrlItem(originalUrl string) (*model.UrlItem, error) {
		shortCode := GenerateShortCode(s.shortUrlLength) 
		urlItem, err := model.NewUrlItem(originalUrl, shortCode)
		if err != nil {
			return nil, err
		}
		err = s.urlRepository.CreateUrl(*urlItem)
		if err != nil {
			return nil, err
		}
		return urlItem, nil
}

func (s *ShortenUrlService) tryGetUrlItem(originalUrl string) (*model.UrlItem, error) {
	var err error
	for range maxCollisionAttempts {
		urlItem, err := s.getUrlItem(originalUrl)
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

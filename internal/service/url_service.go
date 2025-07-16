package service

import (
	"fmt"

	"github.com/oegegr/shortener/internal/model"
	"github.com/oegegr/shortener/internal/repository"
)

type UrlShortner interface {
	GetShortUrl(originalUrl string) (string, error)
	GetOriginalUrl(shortUrl string) (string, error)
}

type ShortenUrlService struct {
	urlRepository  repository.UrlRepository
	urlDomain string
}

func NewShortnerService(repository repository.UrlRepository, domain string) *ShortenUrlService {
	return &ShortenUrlService{urlRepository: repository, urlDomain: domain}
}

func (s *ShortenUrlService) GetShortUrl(originalUrl string) (string, error) {
	shortCode, err := GenerateShortCode(originalUrl)
	if err != nil {
		return "", err
	}

	urlItem, err := model.NewUrlItem(originalUrl, shortCode)
	if err != nil {
		return "", err
	}

	err = s.urlRepository.CreateUrl(*urlItem)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", s.urlDomain, shortCode), nil
}

func (s *ShortenUrlService) GetOriginalUrl(shortCode string) (string, error) {
	urlItem, err := s.urlRepository.FindUrlById(shortCode)
	if err != nil {
		return "", err
	}

	return urlItem.Url, nil
}

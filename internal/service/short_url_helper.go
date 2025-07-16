package service

import (
	"fmt"
	"net/url"
	"strings"
)

func GenerateShortCode(originaUrl string) (string, error) {
	return "foo", nil
}

func GetShortCode(shortUrl string) (string, error) {
	parsedUrl, err := url.ParseRequestURI(shortUrl)
	if err != nil {
		return "", err
	}

	splitedPath := strings.Split(strings.Trim(parsedUrl.Path, "/"), "/")
	if len(splitedPath) == 0 {
		return "", err
	}

	return splitedPath[0], nil
}

func GetShortUrl(shortDomain string, shortCode string) string {
	return fmt.Sprintf("%s/%s", shortDomain, shortCode)
}

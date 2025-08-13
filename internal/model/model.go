package model

import (
	"errors"
)

var (
	ErrInvalidURL = errors.New("invalid URL format")
	ErrEmptyCode  = errors.New("short code cannot be empty")
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type URLItem struct {
	ID  string `json:"short_id"`
	URL string `json:"original_url"`
}

func NewURLItem(url string, id string) *URLItem {
	return &URLItem{
		URL: url,
		ID:  id,
	}
}

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

type ShortenBatchRequest []BatchRequest

type BatchRequest struct {
	URL string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

type ShortenBatchResponse []BatchResponse

type BatchResponse struct {
	Result string `json:"short_url"`
	CorrelationID string `json:"correlation_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type URLItem struct {
	ShortID string `json:"short_id"`
	URL     string `json:"original_url"`
}

func NewURLItem(url string, id string) *URLItem {
	return &URLItem{
		URL:     url,
		ShortID: id,
	}
}

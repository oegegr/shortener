package model

import (
	"errors"
)

var (
	ErrInvalidURL = errors.New("invalid URL format")
	ErrEmptyCode  = errors.New("short code cannot be empty")
)

type UrlItem struct {
	ID  string
	URL string
}


func NewURLItem(url string, id string) (*UrlItem) {
	return &UrlItem{
		URL: url,
		ID:  id,
	}
}

package model

import (
	"errors"
)

var (
	ErrInvalidURL = errors.New("invalid URL format")
	ErrEmptyCode  = errors.New("short code cannot be empty")
)

type URLItem struct {
	ID  string
	URL string
}

func NewURLItem(url string, id string) *URLItem {
	return &URLItem{
		URL: url,
		ID:  id,
	}
}

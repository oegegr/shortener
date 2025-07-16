package model

import (
	"errors"
	"net/url"
)

var (
	ErrInvalidURL = errors.New("invalid URL format")
	ErrEmptyCode  = errors.New("short code cannot be empty")
)

type UrlItem struct {
	Id string 
	Url string 
}


func (s *UrlItem) Validate() error {
	if _, err := url.ParseRequestURI(s.Url); err != nil {
		return ErrInvalidURL
	}
	if len(s.Id) == 0 {
		return ErrEmptyCode
	}
	return nil
}

func NewUrlItem(url string, id string) (*UrlItem, error) {
	su := &UrlItem{
		Url: url,
		Id:   id,
	}
	return su, su.Validate()
}

package repository

import (
	"errors"

	"github.com/oegegr/shortener/internal/model"
)

var (
	ErrRepoNotFound      = errors.New("item not found")
	ErrRepoAlreadyExists = errors.New("item already exists")
)

type URLRepository interface {
	CreateURL(urlItem model.URLItem) error
	FindURLByID(id string) (*model.URLItem, error)
	Exists(id string) bool
}

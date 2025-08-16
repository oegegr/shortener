package repository

import (
	"context"
	"errors"

	"github.com/oegegr/shortener/internal/model"
)

var (
	ErrRepoNotFound      = errors.New("item not found")
	ErrRepoAlreadyExists = errors.New("item already exists")
)

type URLRepository interface {
	CreateURL(ctx context.Context, urlItem []model.URLItem) error
	FindURLByID(ctx context.Context, id string) (*model.URLItem, error)
	Exists(ctx context.Context, id string) bool
}

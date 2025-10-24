package repository

import (
	"context"
	"errors"

	"github.com/oegegr/shortener/internal/model"
)

var (
	ErrRepoNotFound      = errors.New("item not found")
	ErrRepoURLAlreadyExists = errors.New("url already exists")
	ErrRepoShortIDAlreadyExists = errors.New("short id already exists")
)

type URLRepository interface {
	Ping(ctx context.Context) error
	CreateURL(ctx context.Context, urlItem []model.URLItem) error
	DeleteURL(ctx context.Context, ids []string) error
	FindURLByID(ctx context.Context, id string) (*model.URLItem, error)
	FindURLByURL(ctx context.Context, id string) (*model.URLItem, error)
	FindURLByUser(ctx context.Context, userID string) ([]model.URLItem, error)
	Exists(ctx context.Context, id string) bool
}

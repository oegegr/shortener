// Package repository содержит интерфейс и константы для работы с репозиторием URL-адресов.
package repository

import (
	"context"
	"errors"

	"github.com/oegegr/shortener/internal/model"
)

// ErrRepoNotFound представляет ошибку, которая возникает при отсутствии элемента в репозитории.
var ErrRepoNotFound = errors.New("item not found")

// ErrRepoURLAlreadyExists представляет ошибку, которая возникает при попытке создать уже существующий URL-адрес.
var ErrRepoURLAlreadyExists = errors.New("url already exists")

// ErrRepoShortIDAlreadyExists представляет ошибку, которая возникает при попытке создать уже существующий сокращенный идентификатор.
var ErrRepoShortIDAlreadyExists = errors.New("short id already exists")

// URLRepository представляет интерфейс для работы с репозиторием URL-адресов.
type URLRepository interface {
	// Ping проверяет подключение к репозиторию.
	Ping(ctx context.Context) error
	// CreateURL создает новые URL-адреса в репозитории.
	CreateURL(ctx context.Context, urlItem []model.URLItem) error
	// DeleteURL удаляет URL-адреса из репозитория.
	DeleteURL(ctx context.Context, ids []string) error
	// FindURLByID находит URL-адрес в репозитории по идентификатору URL-адреса.
	FindURLByID(ctx context.Context, id string) (*model.URLItem, error)
	// FindURLByURL находит URL-адрес в репозитории по URL-адресу.
	FindURLByURL(ctx context.Context, id string) (*model.URLItem, error)
	// FindURLByUser находит URL-адреса в репозитории по идентификатору пользователя.
	FindURLByUser(ctx context.Context, userID string) ([]model.URLItem, error)
	// Exists проверяет, существует ли URL-адрес в репозитории.
	Exists(ctx context.Context, id string) bool
}

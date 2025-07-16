package repository

import (
	"errors"
	"sync"

	"github.com/oegegr/shortener/internal/model"
)

var (
	ErrRepoNotFound      = errors.New("item not found")
	ErrRepoAlreadyExists = errors.New("item already exists")
)

type URLRepository interface {
	CreateURL(urlItem model.URLItem) error
	FindURLById(id string) (*model.URLItem, error)
	Exists(id string) bool
}

type InMemoryURLRepository struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewInMemoryURLRepository() *InMemoryURLRepository {
	return &InMemoryURLRepository{
		data: make(map[string]string),
	}
}

func (repo *InMemoryURLRepository) CreateURL(urlItem model.URLItem) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	_, ok := repo.data[urlItem.ID]
	if ok {
		return ErrRepoAlreadyExists
	}
	repo.data[urlItem.ID] = urlItem.URL
	return nil
}

func (repo *InMemoryURLRepository) FindURLById(id string) (*model.URLItem, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	url, ok := repo.data[id]
	if !ok {
		return nil, ErrRepoNotFound
	}
	return model.NewURLItem(url, id), nil
}

func (repo *InMemoryURLRepository) Exists(id string) bool {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	_, ok := repo.data[id]
	return ok
}

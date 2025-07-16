package repository 

import (
    "errors"
    "sync"
	"github.com/oegegr/shortener/internal/model"
)

var (
	ErrRepoNotFound = errors.New("item not found")
	ErrRepoAlreadyExists = errors.New("item already exists")
)

type UrlRepository interface {
    CreateUrl(urlItem model.UrlItem) error
    FindUrlById(id string) (*model.UrlItem, error)
    Exists(id string) bool
}

type InMemoryUrlRepository struct {
    mu   sync.RWMutex
    data map[string]string
}

func NewInMemoryUrlRepository() *InMemoryUrlRepository {
    return &InMemoryUrlRepository{
        data: make(map[string]string),
    }
}

func (repo *InMemoryUrlRepository) CreateUrl(urlItem model.UrlItem) error {
    repo.mu.Lock()
    defer repo.mu.Unlock()
    _, ok := repo.data[urlItem.Id] 
    if ok {
        return ErrRepoAlreadyExists
    }
    repo.data[urlItem.Id] = urlItem.Url
    return nil
}

func (repo *InMemoryUrlRepository) FindUrlById(id string) (*model.UrlItem, error) {
    repo.mu.RLock()
    defer repo.mu.RUnlock()
    url, ok := repo.data[id]
    if !ok {
        return nil, ErrRepoNotFound 
    }
    urlItem, err := model.NewUrlItem(url, id)  
    if err != nil {
        return nil, err 

    }
    return urlItem, nil
}

func (repo *InMemoryUrlRepository) Exists(id string) bool {
    repo.mu.RLock()
    defer repo.mu.RUnlock()
    _, ok := repo.data[id]
    return ok
}


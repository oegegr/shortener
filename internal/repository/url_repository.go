package repository

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"go.uber.org/zap"
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

type InMemoryURLRepository struct {
	mu   sync.RWMutex
	data map[string]string
	fileStoragePath string
    logger zap.SugaredLogger
}

func NewInMemoryURLRepository(fileStoragePath string, logger zap.SugaredLogger) *InMemoryURLRepository {
	storage := &InMemoryURLRepository{
		fileStoragePath: fileStoragePath,
		logger: logger,
	}
	err := storage.loadData()
	if err != nil {
		panic("Failed to create InMemory repository with error: " + err.Error())
		

	}
	return storage
}

func (repo *InMemoryURLRepository) CreateURL(urlItem model.URLItem) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	_, ok := repo.data[urlItem.ID]
	if ok {
		return ErrRepoAlreadyExists
	}

	repo.data[urlItem.ID] = urlItem.URL

	err := repo.saveData()

	if err != nil {
		return err
	}

	return nil
}

func (repo *InMemoryURLRepository) FindURLByID(id string) (*model.URLItem, error) {
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

func (repo *InMemoryURLRepository) loadData() error {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	file, err := os.OpenFile(repo.fileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
    if err != nil {
        return err
    }
	defer file.Close()

	var items []model.URLItem
	json.NewDecoder(file).Decode(&items)

	repo.logger.Debugln("Load UrlItems", items)
	data := make(map[string]string, len(items))
	for _, item := range items {
		data[item.ID] = item.URL
	} 
	repo.data = data

	return nil

}


func (repo *InMemoryURLRepository) saveData() error {
	file, err := os.OpenFile(repo.fileStoragePath, os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        return err
    }
	defer file.Close()

	var items []model.URLItem
	for id, url := range repo.data {
		repo.logger.Debugln("Create UrlItem", id, url)
		items = append(items, *model.NewURLItem(url, id))
	}

	return json.NewEncoder(file).Encode(items)
}

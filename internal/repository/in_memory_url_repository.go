package repository

import (
	"encoding/json"
	"os"
	"sync"
	"context"

	"github.com/oegegr/shortener/internal/model"
	"go.uber.org/zap"
)

type InMemoryURLRepository struct {
	mu              sync.RWMutex
	shortIdMap      map[string]string
	urlMap          map[string]string
	fileStoragePath string
	logger          zap.SugaredLogger
}

func NewInMemoryURLRepository(fileStoragePath string, logger zap.SugaredLogger) *InMemoryURLRepository {
	storage := &InMemoryURLRepository{
		fileStoragePath: fileStoragePath,
		logger:          logger,
	}
	err := storage.loadData()
	if err != nil {
		panic("Failed to create InMemory repository with error: " + err.Error())

	}
	return storage
}

func (repo *InMemoryURLRepository) CreateURL(ctx context.Context, items []model.URLItem) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, item := range items {
		_, ok := repo.shortIdMap[item.ShortID]
		if ok {
			return ErrRepoShortIDAlreadyExists
		}

		_, ok = repo.urlMap[item.URL]
		if ok {
			return ErrRepoURLAlreadyExists
		}
	}


	for _, item := range items {
		repo.shortIdMap[item.ShortID] = item.URL
		repo.urlMap[item.URL] = item.ShortID
	}

	err := repo.saveData()

	if err != nil {
		return err
	}

	return nil
}

func (repo *InMemoryURLRepository) FindURLByID(ctx context.Context, id string) (*model.URLItem, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	url, ok := repo.shortIdMap[id]
	if !ok {
		return nil, ErrRepoNotFound
	}
	return model.NewURLItem(url, id), nil
}

func (repo *InMemoryURLRepository) FindURLByURL(ctx context.Context, url string) (*model.URLItem, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	id, ok := repo.urlMap[url]
	if !ok {
		return nil, ErrRepoNotFound
	}
	return model.NewURLItem(url, id), nil
}

func (repo *InMemoryURLRepository) Exists(ctx context.Context, id string) bool {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	_, ok := repo.shortIdMap[id]
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
	shortIDs := make(map[string]string, len(items))
	urls := make(map[string]string, len(items))
	for _, item := range items {
		shortIDs[item.ShortID] = item.URL
		urls[item.URL] = item.ShortID
	}
	repo.shortIdMap = shortIDs
	repo.urlMap = urls

	return nil

}

func (repo *InMemoryURLRepository) saveData() error {
	file, err := os.OpenFile(repo.fileStoragePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var items []model.URLItem
	for id, url := range repo.shortIdMap {
		repo.logger.Debugln("Create UrlItem", id, url)
		items = append(items, *model.NewURLItem(url, id))
	}

	return json.NewEncoder(file).Encode(items)
}
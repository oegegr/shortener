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
	shortIDMap      map[string]string
	urlMap          map[string]string
	fileStoragePath string
	logger          zap.SugaredLogger
	persistent      bool
}

func NewInMemoryURLRepository(fileStoragePath string, logger zap.SugaredLogger) (*InMemoryURLRepository, error) {
	items, err := loadData(fileStoragePath, logger) 
	if err != nil {
		logger.Fatal("Failed to create InMemory repository with error: %w", err.Error())
		return nil, err
	}

	shortIDs := make(map[string]string, len(items))
	urls := make(map[string]string, len(items))
	for _, item := range items {
		shortIDs[item.ShortID] = item.URL
		urls[item.URL] = item.ShortID
	}

	storage := &InMemoryURLRepository{
		fileStoragePath: fileStoragePath,
		logger:          logger,
		persistent: fileStoragePath != "",
		urlMap: urls,
		shortIDMap: shortIDs,
	}

	return storage, nil
}

func (repo *InMemoryURLRepository) Ping(ctx context.Context) error {
	return nil
}

func (repo *InMemoryURLRepository) CreateURL(ctx context.Context, items []model.URLItem) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, item := range items {
		_, ok := repo.shortIDMap[item.ShortID]
		if ok {
			return ErrRepoShortIDAlreadyExists
		}

		_, ok = repo.urlMap[item.URL]
		if ok {
			return ErrRepoURLAlreadyExists
		}
	}


	for _, item := range items {
		repo.shortIDMap[item.ShortID] = item.URL
		repo.urlMap[item.URL] = item.ShortID
	}

	if repo.persistent {
		err := repo.saveData()
		if err != nil {
			return err
		}
	}


	return nil
}

func (repo *InMemoryURLRepository) FindURLByID(ctx context.Context, id string) (*model.URLItem, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	url, ok := repo.shortIDMap[id]
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
	_, ok := repo.shortIDMap[id]
	return ok
}

func (repo *InMemoryURLRepository) saveData() error {
	file, err := os.OpenFile(repo.fileStoragePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var items []model.URLItem
	for id, url := range repo.shortIDMap {
		repo.logger.Debugln("Create UrlItem", id, url)
		items = append(items, *model.NewURLItem(url, id))
	}

	return json.NewEncoder(file).Encode(items)
}

func loadData(fileStoragePath string, logger zap.SugaredLogger) ([]model.URLItem, error) {
	var items []model.URLItem
	if fileStoragePath != "" {
		file, err := os.OpenFile(fileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		json.NewDecoder(file).Decode(&items)
		logger.Debugln("Load UrlItems", items)
		return items, nil
	}
	return []model.URLItem{}, nil 
}

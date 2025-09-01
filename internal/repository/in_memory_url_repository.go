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
	shortIDMap      map[string]model.URLItem
	urlMap          map[string]model.URLItem
	userMap         map[string][]model.URLItem
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

	shortIDs := make(map[string]model.URLItem, len(items))
	urls := make(map[string]model.URLItem, len(items))
	users := make(map[string][]model.URLItem)

	for _, item := range items {
		shortIDs[item.ShortID] = item
		urls[item.URL] = item

		var userItems []model.URLItem
		userItems, ok := users[item.UserID]
		if !ok {
			userItems = []model.URLItem{}
		}

		users[item.UserID] = append(userItems, item)
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
		repo.shortIDMap[item.ShortID] = item
		repo.urlMap[item.URL] = item
		repo.userMap[item.ShortID] = append(repo.userMap[item.ShortID], item)
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
	item, ok := repo.shortIDMap[id]
	if !ok {
		return nil, ErrRepoNotFound
	}
	return model.NewURLItem(item.URL, id, item.UserID), nil
}

func (repo *InMemoryURLRepository) FindURLByURL(ctx context.Context, url string) (*model.URLItem, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	item, ok := repo.urlMap[url]
	if !ok {
		return nil, ErrRepoNotFound
	}
	return model.NewURLItem(url, item.ShortID, item.UserID), nil
}

func (repo *InMemoryURLRepository) FindURLByUser(ctx context.Context, userID string) ([]model.URLItem, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	items, ok := repo.userMap[userID]
	if !ok {
		return nil, ErrRepoNotFound
	}
	return items, nil
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
	for id, item := range repo.shortIDMap {
		repo.logger.Debugln("Create UrlItem", id, item.URL)
		items = append(items, *model.NewURLItem(item.URL, id, item.UserID))
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

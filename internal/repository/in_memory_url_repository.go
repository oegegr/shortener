// Package repository содержит реализацию репозитория для работы с URL-адресами в памяти.
package repository

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/oegegr/shortener/internal/model"

	"github.com/samber/lo/mutable"
	"go.uber.org/zap"
)

// InMemoryURLRepository представляет репозиторий для работы с URL-адресами в памяти.
type InMemoryURLRepository struct {
	// mu представляет mutex для синхронизации доступа к данным.
	mu sync.RWMutex
	// shortIDMap представляет карту сокращенных идентификаторов URL-адресов.
	shortIDMap map[string]model.URLItem
	// urlMap представляет карту URL-адресов.
	urlMap map[string]model.URLItem
	// userMap представляет карту URL-адресов пользователя.
	userMap map[string][]model.URLItem
	// fileStoragePath представляет путь к файлу для хранения данных.
	fileStoragePath string
	// logger представляет логгер для записи сообщений.
	logger zap.SugaredLogger
	// persistent представляет флаг, указывающий, следует ли хранить данные в файле.
	persistent bool
}

// NewInMemoryURLRepository возвращает новый экземпляр InMemoryURLRepository.
// Эта функция принимает путь к файлу для хранения данных и логгер.
func NewInMemoryURLRepository(fileStoragePath string, logger zap.SugaredLogger) (*InMemoryURLRepository, error) {
	items, err := loadData(fileStoragePath, logger)
	if err != nil {
		logger.Fatal("Failed to create InMemory repository with error: %w", err.Error())
		return nil, err
	}

	shortIDs := make(map[string]model.URLItem, len(items))
	urls := make(map[string]model.URLItem, len(items))
	users := make(map[string][]model.URLItem, len(items))

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
		persistent:      fileStoragePath != "",
		urlMap:          urls,
		shortIDMap:      shortIDs,
		userMap:         users,
	}

	return storage, nil
}

// Ping проверяет подключение к репозиторию.
func (repo *InMemoryURLRepository) Ping(ctx context.Context) error {
	return nil
}

// CreateURL создает новые URL-адреса в репозитории.
// Эта функция принимает список элементов URL-адресов для создания.
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
		repo.userMap[item.UserID] = append(repo.userMap[item.UserID], item)
	}

	if repo.persistent {
		err := repo.saveData()
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteURL удаляет URL-адреса из репозитория.
// Эта функция принимает список идентификаторов URL-адресов для удаления.
func (repo *InMemoryURLRepository) DeleteURL(ctx context.Context, ids []string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, id := range ids {
		item, ok := repo.shortIDMap[id]

		if !ok {
			return ErrRepoNotFound
		}

		item.IsDeleted = true
		repo.shortIDMap[item.ShortID] = item
		repo.urlMap[item.URL] = item
		userItems := repo.userMap[item.UserID]

		for idx, userItem := range userItems {
			if userItem.ShortID == id {
				userItems[idx].IsDeleted = true
				break
			}
		}

	}

	if repo.persistent {
		err := repo.saveData()
		if err != nil {
			return err
		}
	}

	return nil
}

// FindURLByID находит URL-адрес в репозитории по идентификатору URL-адреса.
// Эта функция принимает идентификатор URL-адреса для поиска.
func (repo *InMemoryURLRepository) FindURLByID(ctx context.Context, id string) (*model.URLItem, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	item, ok := repo.shortIDMap[id]
	if !ok || item.IsDeleted {
		return nil, ErrRepoNotFound
	}
	return model.NewURLItem(item.URL, id, item.UserID, item.IsDeleted), nil
}

// FindURLByURL находит URL-адрес в репозитории по URL-адресу.
// Эта функция принимает URL-адрес для поиска.
func (repo *InMemoryURLRepository) FindURLByURL(ctx context.Context, url string) (*model.URLItem, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	item, ok := repo.urlMap[url]
	if !ok || item.IsDeleted {
		return nil, ErrRepoNotFound
	}
	return model.NewURLItem(url, item.ShortID, item.UserID, item.IsDeleted), nil
}

// FindURLByUser находит URL-адреса в репозитории по идентификатору пользователя.
// Эта функция принимает идентификатор пользователя для поиска.
func (repo *InMemoryURLRepository) FindURLByUser(ctx context.Context, userID string) ([]model.URLItem, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	items, ok := repo.userMap[userID]

	if !ok {
		return nil, ErrRepoNotFound
	}

	result := mutable.Filter(items, func(item model.URLItem) bool { return !item.IsDeleted })

	if len(result) == 0 {
		return nil, ErrRepoNotFound
	}
	return result, nil
}

// Exists проверяет, существует ли URL-адрес в репозитории.
// Эта функция принимает идентификатор URL-адреса для проверки.
func (repo *InMemoryURLRepository) Exists(ctx context.Context, id string) bool {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	_, ok := repo.shortIDMap[id]
	return ok
}

// saveData сохраняет данные в файле.
func (repo *InMemoryURLRepository) saveData() error {
	file, err := os.OpenFile(repo.fileStoragePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var items []model.URLItem
	for _, item := range repo.shortIDMap {
		repo.logger.Debugln("Create UrlItem", item.ShortID, item.URL)
		items = append(items, *model.NewURLItem(item.URL, item.ShortID, item.UserID, item.IsDeleted))
	}

	return json.NewEncoder(file).Encode(items)
}

// loadData загружает данные из файла.
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

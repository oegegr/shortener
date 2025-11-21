package internal

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	pkghttp "github.com/oegegr/shortener/pkg/http"
	"github.com/oegegr/shortener/internal/config"
	"github.com/oegegr/shortener/internal/config/db"
	sugar "github.com/oegegr/shortener/internal/config/logger"
	"github.com/oegegr/shortener/internal/repository"
	"github.com/oegegr/shortener/internal/service"
	"go.uber.org/zap"
)

// ShotenerAppBuilder - билдер для создания ShortenerApp
type ShotenerAppBuilder struct {
	cfg *config.Config
}

// NewShortenerAppBuilder - конструктор билдера
func NewShortenerAppBuilder(cfg *config.Config) *ShotenerAppBuilder {
	return &ShotenerAppBuilder{cfg}
}

// Build - создает и конфигурирует ShortenerApp
func (builder *ShotenerAppBuilder) Build(ctx context.Context) (*ShortenerApp, error) {
	logger, err := sugar.NewLogger(*builder.cfg)
	if err != nil {
		return nil, err
	}

	var dbConn *sql.DB
	if builder.cfg.DBConnectionString != "" {
		dbConn, err = db.NewDB(*builder.cfg, logger)
		if err != nil {
			logger.Error("failed to create db connection: %w", err)
			return nil, err
		}
	}

	repo, err := createURLRepository(*builder.cfg, *logger, dbConn)
	if err != nil {
		logger.Error("failed to create repository: %w", err)
		return nil, err
	}

	urlDelStrategy := createURLDeletionStrategy(*logger, repo)

	service := createShortnerService(*builder.cfg, *logger, repo, urlDelStrategy)

	jwtParser := createJWTParser(*builder.cfg, *logger)

	logAudit := createLogAudit(*builder.cfg)

	router := NewShortenerRouter(*logger, service, jwtParser, repo, logAudit)

	server, err := createServer(router, *builder.cfg)
	if err != nil {
		logger.Error("failed to create server: %w", err)
		return nil, err
	}

	stopApp := func(ctx context.Context) {
		logger.Sync()
		dbConn.Close()
		urlDelStrategy.Stop()
	}

	return NewShortenerApp(builder.cfg, server, dbConn, logger, stopApp), nil
}

func createServer(handler http.Handler, cfg config.Config) (pkghttp.Server, error) {
	serverBuilder := pkghttp.NewServerBuilder(cfg.ServerAddress)

	if cfg.EnableHTTPS {
		serverBuilder.WithHTTPS(cfg.TLSCertFile, cfg.TLSKeyFile)
	}

	server, err := serverBuilder.Build(handler)
	if err != nil {
		return nil, err
	}
	return server, nil
}

// createURLRepository - создает репозиторий URL (БД или in-memory)
func createURLRepository(
	c config.Config,
	logger zap.SugaredLogger,
	db *sql.DB,
) (repository.URLRepository, error) {

	if c.DBConnectionString != "" {
		return repository.NewDBURLRepository(db, logger)
	}

	return repository.NewInMemoryURLRepository(c.FileStoragePath, logger)
}

// createURLDeletionStrategy - создает стратегию удаления URL через очередь
func createURLDeletionStrategy(
	logger zap.SugaredLogger,
	repo repository.URLRepository,
) *service.QueueDeletionStrategy {
	workerNum := 5
	taskNum := 1000
	waitTimeout := 1 * time.Second
	return service.NewQueueURLDeletionStrategy(repo, logger, workerNum, taskNum, waitTimeout)
}

// createShortnerService - создает сервис сокращения URL
func createShortnerService(
	c config.Config,
	logger zap.SugaredLogger,
	repo repository.URLRepository,
	urlDelStrategy service.URLDeletionStrategy,
) service.URLShortener {

	return service.NewShortenerService(
		repo,
		c.BaseURL,
		c.ShortURLLength,
		&service.RandomShortCodeProvider{},
		urlDelStrategy,
		logger,
	)
}

// createLogAudit - создает менеджер аудита логов
func createLogAudit(c config.Config) service.LogAuditManager {
	auditors := make([]service.LogAuditor, 0, 2)

	if c.AuditFile != "" {
		auditors = append(auditors, service.NewFileLogAuditor(c.AuditFile))
	}

	if c.AuditURL != "" {
		auditors = append(auditors, service.NewHTTPLogAuditor(c.AuditURL))
	}

	return service.NewDefaultLogAuditManager(auditors)
}

// createJWTParser - создает парсер JWT токенов
func createJWTParser(
	c config.Config,
	logger zap.SugaredLogger,
) service.JWTParser {
	return service.NewJWTParser(c.JWTSecret, logger)
}

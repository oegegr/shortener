package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/oegegr/shortener/internal/config"
	"github.com/oegegr/shortener/internal/config/db"
	"github.com/oegegr/shortener/internal/repository"
	"github.com/oegegr/shortener/internal/service"
	pkghttp "github.com/oegegr/shortener/pkg/http"
	"go.uber.org/zap"
)

// ShotenerAppBuilder - билдер для создания ShortenerApp
type ShotenerAppBuilder struct {
	cfg    *config.Config
	logger *zap.SugaredLogger
}

// NewShortenerAppBuilder - конструктор билдера
func NewShortenerAppBuilder(cfg *config.Config, logger *zap.SugaredLogger) *ShotenerAppBuilder {
	return &ShotenerAppBuilder{cfg, logger}
}

// Build - создает и конфигурирует ShortenerApp
func (b *ShotenerAppBuilder) Build(ctx context.Context) (*ShortenerApp, func(context.Context, *zap.SugaredLogger), error) {
	var err error
	var dbConn *sql.DB
	if b.cfg.DBConnectionString != "" {
		dbConn, err = db.NewDB(*b.cfg, b.logger)
		if err != nil {
			b.logger.Error("failed to create db connection: %w", err)
			return nil, nil, fmt.Errorf("failed to create db connection: %v", err)
		}
	}

	repo, err := createURLRepository(*b.cfg, *b.logger, dbConn)
	if err != nil {
		b.logger.Error("failed to create repository: %w", err)
		return nil, nil, err
	}

	urlDelStrategy := createURLDeletionStrategy(*b.logger, repo)

	service := createShortnerService(*b.cfg, *b.logger, repo, urlDelStrategy)

	jwtParser := createJWTParser(*b.cfg, *b.logger)

	logAudit := createLogAudit(*b.cfg)

	router := NewShortenerRouter(*b.logger, service, jwtParser, repo, logAudit)

	server, err := createServer(router, *b.cfg)
	if err != nil {
		b.logger.Error("failed to create server: %w", err)
		return nil, nil, err
	}

	stopApp := func(stopCtx context.Context, logger *zap.SugaredLogger) {
		var stopErrors []error

		logger.Info("Starting application cleanup...")

		if dbConn != nil {
			b.logger.Info("Closing database connection...")
			if err := dbConn.Close(); err != nil {
				stopErrors = append(stopErrors, fmt.Errorf("failed to close database: %w", err))
			}
		}

		b.logger.Info("Stoping urlDelStrategy...")
		urlDelStrategy.Stop()

		if server != nil {
			logger.Info("Shutting down HTTP server...")
			shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			if err := server.Stop(shutdownCtx); err != nil {
				stopErrors = append(stopErrors, fmt.Errorf("HTTP server shutdown failed: %w", err))
			}
		}

		logger.Info("Syncing logger...")
		if err := logger.Sync(); err != nil {
			logger.Debugf("Logger sync warning: %v", err)
		}

		if len(stopErrors) > 0 {
			logger.Fatal("failed to stop gracefuly application with errors: %v", errors.Join(stopErrors...))
		}
	}

	return NewShortenerApp(b.cfg, server, dbConn, b.logger), stopApp, nil
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

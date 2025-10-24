package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/oegegr/shortener/internal"
	"github.com/oegegr/shortener/internal/config"
	"github.com/oegegr/shortener/internal/config/db"
	sugar "github.com/oegegr/shortener/internal/config/logger"
	"github.com/oegegr/shortener/internal/repository"
	"github.com/oegegr/shortener/internal/service"
	"go.uber.org/zap"
)

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
func createURLDeletionStrategy(
	logger zap.SugaredLogger,
	repo repository.URLRepository,
) *service.QueueDeletionStrategy {
	workerNum := 5
	taskNum := 1000
	waitTimeout, _  := time.ParseDuration("PT1S")
	return service.NewQueueURLDeletionStrategy(repo, logger, workerNum, taskNum, waitTimeout)
}

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

func createLogAudit(c config.Config) service.LogAuditManager {
	auditors := make([]service.LogAuditor, 0, 2)

	if c.AuditFile != "" {
		auditors = append(auditors, service.NewFileLogAuditor(c.AuditFile))
	}

	if c.AuditURL != "" {
		auditors = append(auditors, service.NewHttpLogAuditor(c.AuditURL))
	}

	return service.NewDefaultLogAuditManager(auditors)
}

func createJWTParser(
	c config.Config,
	logger zap.SugaredLogger,
) service.JWTParser {
	return service.NewJWTParser(c.JWTSecret, logger)
}

func main() {
	c := config.NewConfig()

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	logger, err := sugar.NewLogger(c)
	defer logger.Sync()

	if err != nil {
		os.Exit(1)
	}

	var dbConn *sql.DB
	if c.DBConnectionString != "" {
		dbConn, err = db.NewDB(c, logger)
		if err != nil {
			logger.Error("failed to create db connection: %w", err)
			os.Exit(1)
		}

		defer dbConn.Close()
	}

	repo, err := createURLRepository(c, *logger, dbConn)
	if err != nil {
		logger.Error("failed to create repository: %w", err)
		os.Exit(1)
	}


	urlDelStrategy := createURLDeletionStrategy(*logger, repo) 
	defer urlDelStrategy.Stop()

	service := createShortnerService(c, *logger, repo, urlDelStrategy)

	jwtParser := createJWTParser(c, *logger)

	logAudit := createLogAudit(c)

	router := internal.NewShortenerRouter(*logger, service, jwtParser, repo, logAudit)

	go func() {
		logger.Infoln("Server starting")
		if err := http.ListenAndServe(c.ServerAddress, router); err != nil && err != http.ErrServerClosed {
			logger.Infoln("Server stopped: %v\n", err)
			stop()
		}
	}()

	<-ctx.Done()
}

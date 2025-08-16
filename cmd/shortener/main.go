package main

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/oegegr/shortener/internal/config"
	"github.com/oegegr/shortener/internal/handler"
	"github.com/oegegr/shortener/internal/middleware"
	"github.com/oegegr/shortener/internal/repository"
	"github.com/oegegr/shortener/internal/service"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func createLogger(c config.Config) zap.SugaredLogger {
	var sugar zap.SugaredLogger
	level, err := zapcore.ParseLevel(c.LogLevel)

	if err != nil {
		panic(err)
	}

	logCfg := zap.NewDevelopmentConfig()
	logCfg.Level = zap.NewAtomicLevelAt(level)
	logger, err := logCfg.Build()

	if err != nil {
		panic("failed to create logger: " + err.Error())
	}

	sugar = *logger.Sugar()

	return sugar
}

func createDB(c config.Config) *sql.DB {
	if c.DBConnectionString == "" {
		return nil
	}
	db, err := sql.Open("pgx", c.DBConnectionString)

	if err != nil {
		panic("failed to create db connection: " + err.Error())
	}

	return db
}

func createURLRepository(
	c config.Config, 
	ctx context.Context, 
	logger zap.SugaredLogger,
	db *sql.DB,
	) repository.URLRepository {

	if c.DBConnectionString != "" {
		return repository.NewDBURLRepository(ctx, db, logger)
	} 

	return repository.NewInMemoryURLRepository(c.FileStoragePath, logger)
}

func createShortnerService(
	ctx context.Context, 
	c config.Config,
	logger zap.SugaredLogger, 
	repo repository.URLRepository,
	) service.URLShortener {
	return service.NewShortenerService(
		repo,
		c.BaseURL,
		c.ShortURLLength,
		&service.RandomShortCodeProvider{},
		ctx,
		logger,
	)
}

func main() {
	c := config.NewConfig()

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	db := createDB(c)
	defer db.Close()

	logger := createLogger(c)
	defer logger.Sync()

	repo := createURLRepository(c, ctx, logger, db)
	service := createShortnerService(ctx, c, logger, repo)

	shortenerHandler := handler.NewShortenerHandler(service)
	pingHandler := handler.NewPingHandler(db)

	router := chi.NewRouter()
	router.Use(middleware.ZapLogger(logger))
	typesToGzip := []string{"application/json", "text/html"}
	router.Use(middleware.GzipMiddleware(typesToGzip))
	router.Get("/ping", pingHandler.Ping)
	router.Post("/api/shorten", shortenerHandler.APIShortenURL)
	router.Post("/*", shortenerHandler.ShortenURL)
	router.Get("/{short_url}", shortenerHandler.RedirectToOriginalURL)

	go func() {
		logger.Infoln("Server starting")
		if err := http.ListenAndServe(c.ServerAddress, router); err != nil && err != http.ErrServerClosed {
			logger.Infoln("Server stopped: %v\n", err)
			stop()
		}
	}()

	<-ctx.Done()
}

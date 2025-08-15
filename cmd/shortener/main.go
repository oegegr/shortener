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

func createDb(c config.Config) *sql.DB {
	db, err := sql.Open("pgx", c.DBConnectionString)

	if err != nil {
		panic("failed to create db connection: " + err.Error())
	}

	return db
}

func main() {
	c := config.NewConfig()

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	db := createDb(c)
	defer db.Close()

	logger := createLogger(c)
	defer logger.Sync()

	urlRepository := repository.NewInMemoryURLRepository(c.FileStoragePath, logger)
	urlService := service.NewShortenerService(
		urlRepository,
		c.BaseURL,
		c.ShortURLLength,
		&service.RandomShortCodeProvider{},
		ctx,
		logger,
	)
	shortenerHandler := handler.NewShortenerHandler(urlService)
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

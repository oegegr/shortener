package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
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
		panic("Failed to create logger: " + err.Error())
	}

	defer logger.Sync()

	sugar = *logger.Sugar()

	return sugar
}

func main() {
	c := config.NewConfig()

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	logger := createLogger(*c)

	urlRepository := repository.NewInMemoryURLRepository(c.FileStoragePath, logger)
	urlService := service.NewShortenerService(
		urlRepository,
		c.BaseURL,
		c.ShortURLLength,
		&service.RandomShortCodeProvider{},
	    ctx,
		logger,
	)
	ShortenerHandler := handler.NewShortenerHandler(urlService)

	router := chi.NewRouter()
	router.Use(middleware.ZapLogger(logger))
	typesToGzip := []string{"application/json", "text/html"}
	router.Use(middleware.GzipMiddleware(typesToGzip))
	router.Post("/api/shorten", ShortenerHandler.APIShortenURL)
	router.Post("/*", ShortenerHandler.ShortenURL)
	router.Get("/{short_url}", ShortenerHandler.RedirectToOriginalURL)

	go func() {
		logger.Infoln("Server starting")
		if err := http.ListenAndServe(c.ServerAddress, router); err != nil && err != http.ErrServerClosed {
			logger.Infoln("Server stopped: %v\n", err)
			stop()
		}
	}()

	<-ctx.Done()
}

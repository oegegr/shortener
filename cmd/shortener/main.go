package main

import (
	"context"
	"fmt"
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
		panic(err)
	}

	defer logger.Sync()

	sugar = *logger.Sugar()

	return sugar
}

func main() {
	c := config.NewConfig()

	logger := createLogger(*c)

	urlRepository := repository.NewInMemoryURLRepository(c.FileStoragePath, logger)
	urlService := service.NewShortenerService(
		urlRepository,
		c.BaseURL,
		c.ShortURLLength,
		&service.RandomShortCodeProvider{})
	ShortenerHandler := handler.NewShortenerHandler(urlService)

	router := chi.NewRouter()
	router.Use(middleware.ZapLogger(logger))
	typesToGzip := []string{"application/json", "text/html"}
	router.Use(middleware.GzipMiddleware(typesToGzip))
	router.Post("/api/shorten", ShortenerHandler.APIShortenURL)
	router.Post("/*", ShortenerHandler.ShortenURL)
	router.Get("/{short_url}", ShortenerHandler.RedirectToOriginalURL)

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	go func() {
		fmt.Println("Server starting")
		if err := http.ListenAndServe(c.ServerAddress, router); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server stopped: %v\n", err)
			stop()
		}
	}()

	<-ctx.Done()
}

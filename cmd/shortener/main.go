package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/oegegr/shortener/internal/config"
	"github.com/oegegr/shortener/internal/handler"
	"github.com/oegegr/shortener/internal/repository"
	"github.com/oegegr/shortener/internal/service"
)

func main() {
	c := config.NewConfig()
	urlRepository := repository.NewInMemoryURLRepository()
	urlService := service.NewShortenerService(
		urlRepository,
		c.BaseURL,
		c.ShortURLLength,
		&service.RandomShortCodeProvider{})
	ShortenerHandler := handler.NewShortenerHandler(urlService)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
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

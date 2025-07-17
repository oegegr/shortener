package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/oegegr/shortener/internal/handler"
	"github.com/oegegr/shortener/internal/repository"
	"github.com/oegegr/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	urlRepository := repository.NewInMemoryURLRepository()
	urlService := service.NewShortnerService(urlRepository, "http://127.0.0.1:8080", 8, &service.RandomShortCodeProvider{})
	shortnerHandler := handler.NewShortnerHandler(urlService)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.HandleFunc("/{short_url}", shortnerHandler.RedirectToOriginalURL)
	router.HandleFunc("/", shortnerHandler.ShortenURL)

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	go func() {
		fmt.Println("Server starting")
		if err := http.ListenAndServe(`:8080`, router); err != nil {
			fmt.Println("Server stopped")
			stop()
		}
	}()

	<-ctx.Done()
}

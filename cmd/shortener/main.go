package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/oegegr/shortener/internal/handler"
	"github.com/oegegr/shortener/internal/repository"
	"github.com/oegegr/shortener/internal/service"
)

func main() {

	urlRepository := repository.NewInMemoryURLRepository()
	urlService := service.NewShortnerService(urlRepository, "http://127.0.0.1:8080", 8)
	shortnerHandler := handler.NewShortnerHandler(*urlService)

	mux := http.NewServeMux()
	mux.HandleFunc("/{short_url}", shortnerHandler.RedirectToOriginalURL)
	mux.HandleFunc("/", shortnerHandler.ShortenURL)

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	go func() {
		fmt.Println("Server starting")
		if err := http.ListenAndServe(`:8080`, mux); err != nil {
			fmt.Println("Server stopped")
			stop()
		}
	}()

	<-ctx.Done()
}

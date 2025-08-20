package internal

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/oegegr/shortener/internal/handler"
	"github.com/oegegr/shortener/internal/middleware"
	"github.com/oegegr/shortener/internal/service"
	"go.uber.org/zap"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func NewShortenerRouter(
	logger zap.SugaredLogger,
	service service.URLShortener,
	db *sql.DB,
) *chi.Mux {

	shortenerHandler := handler.NewShortenerHandler(service)
	pingHandler := handler.NewPingHandler(db)

	router := chi.NewRouter()
	router.Use(middleware.ZapLogger(logger))
	typesToGzip := []string{"application/json", "text/html"}
	router.Use(middleware.GzipMiddleware(typesToGzip))
	router.Get("/ping", pingHandler.Ping)
	router.Post("/api/shorten/batch", shortenerHandler.APIShortenBatchURL)
	router.Post("/api/shorten", shortenerHandler.APIShortenURL)
	router.Post("/*", shortenerHandler.ShortenURL)
	router.Get("/{short_url}", shortenerHandler.RedirectToOriginalURL)

	return router
}

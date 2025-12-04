// Package internal содержит реализацию роутера для приложения.
package internal

import (
	"github.com/go-chi/chi/v5"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/oegegr/shortener/internal/handler"
	"github.com/oegegr/shortener/internal/middleware"
	"github.com/oegegr/shortener/internal/repository"
	"github.com/oegegr/shortener/internal/service"
	"go.uber.org/zap"

	pkgnet "github.com/oegegr/shortener/pkg/net"
)

// NewShortenerRouter возвращает новый экземпляр роутера для приложения.
// Эта функция принимает логгер, сервис сокращения URL-адресов, парсер JWT-токенов, репозиторий URL-адресов и менеджер аудита логов.
func NewShortenerRouter(
	logger zap.SugaredLogger,
	service service.URLShortener,
	jwtParser service.JWTParser,
	repo repository.URLRepository,
	logAudit service.LogAuditManager,
	trustedSubnet *pkgnet.Subnet,
) *chi.Mux {
	shortenerHandler := handler.NewShortenerHandler(service, &middleware.AuthContextUserIDPovider{}, logAudit)
	pingHandler := handler.NewPingHandler(repo)

	router := chi.NewRouter()
	router.Use(middleware.ZapLogger(logger))
	typesToGzip := []string{"application/json", "text/html"}
	trustedSubnetAuthorizer := middleware.TrusteSubnetAuthorizer(trustedSubnet, logger)
	router.Use(
		middleware.ZapLogger(logger),
		middleware.GzipMiddleware(typesToGzip),
		middleware.AuthMiddleware(logger, jwtParser),
	)
	router.Get("/ping", pingHandler.Ping)
	router.Post("/api/shorten/batch", shortenerHandler.APIShortenBatchURL)
	router.Post("/api/shorten", shortenerHandler.APIShortenURL)
	router.Get("/api/user/urls", shortenerHandler.APIUserURL)
	router.Delete("/api/user/urls", shortenerHandler.APIUserBatchDeleteURL)
	router.Post("/*", shortenerHandler.ShortenURL)
	router.Get("/{short_url}", shortenerHandler.RedirectToOriginalURL)
	router.Get("/api/internal/stats", trustedSubnetAuthorizer(shortenerHandler.APIInternalStats))

	return router
}

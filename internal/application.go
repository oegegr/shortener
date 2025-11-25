package internal

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/oegegr/shortener/internal/config"
	pkghttp "github.com/oegegr/shortener/pkg/http"
	"go.uber.org/zap"
)

// ShortenerApp - основное приложение для сокращения URL
type ShortenerApp struct {
	cfg    *config.Config
	server pkghttp.Server
	dbConn *sql.DB
	logger *zap.SugaredLogger
}

// Конструктор для ShortenerApp
func NewShortenerApp(
	cfg *config.Config,
	server pkghttp.Server,
	dbConn *sql.DB,
	logger *zap.SugaredLogger,
) *ShortenerApp {
	return &ShortenerApp{cfg, server, dbConn, logger}
}

// Start - запускает приложение
func (app *ShortenerApp) Start(appCtx context.Context) error {

	serverErr := make(chan error, 1)
	go func() {
		app.logger.Info("Server starting")
		if err := app.server.Start(appCtx); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-appCtx.Done():
		app.logger.Info("Shutdown signal received, stopping application")
		return nil
	}
}

package internal

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/oegegr/shortener/internal/config"
	"go.uber.org/zap"
)

// ShortenerApp - основное приложение для сокращения URL
type ShortenerApp struct {
	cfg     *config.Config
	server  *http.Server
	dbConn  *sql.DB
	logger  *zap.SugaredLogger
	stopApp func(ctx context.Context)
}

// Конструктор для ShortenerApp
func NewShortenerApp(
	cfg *config.Config,
	server *http.Server,
	dbConn *sql.DB,
	logger *zap.SugaredLogger,
	stopApp func(ctx context.Context),
) *ShortenerApp {
	return &ShortenerApp{cfg, server, dbConn, logger, stopApp}
}

// Start - запускает приложение
func (app *ShortenerApp) Start(ctx context.Context) error {
	ctx, stop := context.WithCancel(ctx)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		app.logger.Info("Server starting")
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		app.server.Shutdown(shutdownCtx)
		app.stopApp(ctx)
		return nil
	}
}

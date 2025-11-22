package internal

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/oegegr/shortener/internal/config"
	pkghttp "github.com/oegegr/shortener/pkg/http"
	"go.uber.org/zap"
)

// ShortenerApp - основное приложение для сокращения URL
type ShortenerApp struct {
	cfg     *config.Config
	server  pkghttp.Server
	dbConn  *sql.DB
	logger  *zap.SugaredLogger
	stopApp func(ctx context.Context) error
}

// Конструктор для ShortenerApp
func NewShortenerApp(
	cfg *config.Config,
	server pkghttp.Server,
	dbConn *sql.DB,
	logger *zap.SugaredLogger,
	stopApp func(ctx context.Context) error,
) *ShortenerApp {
	return &ShortenerApp{cfg, server, dbConn, logger, stopApp}
}

// Start - запускает приложение
func (app *ShortenerApp) Start(appCtx context.Context) error {
	// Создаем контекст который отменяется по сигналам завершения
	appCtx, stop := signal.NotifyContext(appCtx,
		syscall.SIGTERM, // Сигнал завершения (kill)
		syscall.SIGINT,  // Сигнал прерывания (Ctrl+C)
		syscall.SIGQUIT, // Сигнал выхода (Ctrl+\)
	)
	defer stop()

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
		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Выполняем graceful shutdown
		if err := app.gracefulShutdown(shutdownCtx); err != nil {
			app.logger.Errorf("Graceful shutdown failed: %v", err)
			return err
		}

		app.logger.Info("Server stopped gracefully")
		return nil
	}
}

// gracefulShutdown выполняет корректное завершение работы
func (app *ShortenerApp) gracefulShutdown(ctx context.Context) error {
	app.logger.Info("Starting graceful shutdown process...")
	if app.stopApp != nil {
		if err := app.stopApp(ctx); err != nil {
			app.logger.Errorf("Error during app cleanup: %v", err)
			return fmt.Errorf("app cleanup failed: %w", err)
		}
	}

	app.logger.Info("Graceful shutdown completed successfully")
	return nil
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oegegr/shortener/internal"
	"github.com/oegegr/shortener/internal/config"
	sugar "github.com/oegegr/shortener/internal/config/logger"
	"go.uber.org/zap"
)

// Переменные для хранения информации о сборке
// Заполняются при сборке через ldflags
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

// Основная функция приложения
func main() {
	printBuildInfo()

	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Printf("Failed to load app config: %v", err)
		os.Exit(1)
	}
	printAppConfig(cfg)

	logger, err := sugar.NewLogger(*cfg)
	if err != nil {
		fmt.Printf("Failed to creat logger: %v", err)
		os.Exit(1)
	}

	// Создаем контекст который отменяется по сигналам завершения
	appCtx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGTERM, // Сигнал завершения (kill)
		syscall.SIGINT,  // Сигнал прерывания (Ctrl+C)
		syscall.SIGQUIT, // Сигнал выхода (Ctrl+\)
	)
	defer stop()

	app, stopApp, err := internal.NewShortenerAppBuilder(cfg, logger).Build(appCtx)
	if err != nil {
		logger.Fatal("failed to create application: %w", err)
		os.Exit(1)
	}
	defer gracefulShutdown(stopApp, logger, 5*time.Second)

	if err := app.Start(appCtx); err != nil {
		logger.Fatal("failed to start application: %w", err)
		os.Exit(1)
	}
}

// Функция для graceful остановки приложения
func gracefulShutdown(
	stopApp func(context.Context, *zap.SugaredLogger),
	logger *zap.SugaredLogger,
	gracefulShutdownTimeout time.Duration) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()
	logger.Info("Starting graceful shutdown process...")
	stopApp(shutdownCtx, logger)
	logger.Sync()
}

// Функция для вывода информации о конфигурации приложения
func printAppConfig(cfg *config.Config) {
	jsonConfig, err := json.MarshalIndent(cfg, "", " ")
	if err != nil {
		fmt.Printf("Failed to print app config: %v", err)
		return
	}
	fmt.Printf("Application Config: %s", jsonConfig)
}

// Функция для вывода информации о сборке
func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/oegegr/shortener/internal"
	"github.com/oegegr/shortener/internal/config"
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
		log.Fatalf("Failed to load app config: %v", err)
		os.Exit(1)
	}
	printAppConfig(cfg)

	appCtx := context.Background()

	app, err := internal.NewShortenerAppBuilder(cfg).Build(appCtx)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Start(ctx); err != nil {
		log.Fatalf("Server failed to start: %v", err)
		os.Exit(1)
	}
}

// Функция для вывода информации о конфигурации приложения
func printAppConfig(cfg *config.Config) {
	jsonConfig, err := json.MarshalIndent(cfg, "", " ")
	if err != nil {
		log.Fatalf("Failed to print app config: %v", err)
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

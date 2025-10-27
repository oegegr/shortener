// Package config содержит реализацию конфигурации приложения.
package config

import (
	"flag"
	"os"
)

// Config представляет структуру конфигурации приложения.
type Config struct {
	// ServerAddress представляет адрес сервера.
	ServerAddress      string
	// BaseURL представляет базовый URL-адрес для сокращенных URL-адресов.
	BaseURL            string
	// ShortURLLength представляет длину сокращенного URL-адреса.
	ShortURLLength     int
	// FileStoragePath представляет путь к файлу для хранения данных.
	FileStoragePath    string
	// DBConnectionString представляет строку подключения к базе данных.
	DBConnectionString string
	// LogLevel представляет уровень логирования.
	LogLevel           string
	// JWTSecret представляет секретный ключ для JWT-токенов.
	JWTSecret          string
	// AuditFile представляет файл для хранения аудит-логов.
	AuditFile          string
	// AuditURL представляет URL-адрес для отправки аудит-логов.
	AuditURL           string
}

// NewConfig возвращает новый экземпляр конфигурации приложения.
// Эта функция парсит флаги командной строки и переменные окружения для инициализации конфигурации.
func NewConfig() Config {
	cfg := Config{}

	flag.StringVar(&cfg.ServerAddress, "a", "127.0.0.1:8080", "address to startup server")
	flag.StringVar(&cfg.BaseURL, "b", "http://127.0.0.1:8080", "domain to use for shrten urls")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file path to save storage")
	flag.StringVar(&cfg.DBConnectionString, "d", "", "database connection string")
	flag.StringVar(&cfg.LogLevel, "l", "DEBUG", "log level")
	flag.StringVar(&cfg.JWTSecret, "s", "jwt-secret-key", "jwt secret key")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "file to keep audit logs")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "URL to pass audit logs")
	flag.IntVar(&cfg.ShortURLLength, "c", 8, "length of generated short url")
	flag.Parse()

	if envServerAddress, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.ServerAddress = envServerAddress
	}

	if envBaseURL, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.BaseURL = envBaseURL
	}

	if fileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.FileStoragePath = fileStoragePath
	}

	if LogLevel, ok := os.LookupEnv("LOG_LEVEL"); ok {
		cfg.LogLevel = LogLevel
	}

	if dbConnectionString, ok := os.LookupEnv("DATABASE_DSN"); ok {
		cfg.DBConnectionString = dbConnectionString
	}

	if jwtSecret, ok := os.LookupEnv("JWT_SECRET"); ok {
		cfg.JWTSecret = jwtSecret
	}

	if auditFile, ok := os.LookupEnv("AUDIT_FILE"); ok {
		cfg.AuditFile = auditFile
	}

	if auditURL, ok := os.LookupEnv("AUDIT_URL"); ok {
		cfg.AuditURL = auditURL
	}

	return cfg
}

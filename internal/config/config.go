package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress      string
	BaseURL            string
	ShortURLLength     int
	FileStoragePath    string
	DBConnectionString string
	LogLevel           string
	JWTSecret          string
	AuditFile          string
	AuditURL           string
}

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

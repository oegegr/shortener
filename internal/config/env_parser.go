// Package config содержит реализацию конфигурации приложения.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// EnvConfigParser парсит конфигурацию из переменных окружения
type EnvConfigParser struct {
	NextParser
}

// EnvConfigParser парсит конфигурацию из переменных окружения
func (e *EnvConfigParser) Parse(cfg *Config) (*Config, error) {
	if envServerAddress, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.ServerAddress = envServerAddress
	}
	if envBaseURL, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.BaseURL = envBaseURL
	}
	if fileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.FileStoragePath = fileStoragePath
	}
	if logLevel, ok := os.LookupEnv("LOG_LEVEL"); ok {
		cfg.LogLevel = logLevel
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
	if envEnableHTTPS, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		if envEnableHTTPS == "true" {
			cfg.EnableHTTPS = true
		}
	}
	if envCertFile, ok := os.LookupEnv("TLS_CERT_FILE"); ok {
		cfg.TLSCertFile = envCertFile
	}
	if envKeyFile, ok := os.LookupEnv("TLS_KEY_FILE"); ok {
		cfg.TLSKeyFile = envKeyFile
	}
	if trustedSubnet, ok := os.LookupEnv("TRUSTED_SUBNET"); ok {
		cfg.TrustedSubnet = trustedSubnet
	}
	if grpcPort, ok := os.LookupEnv("GRPC_PORT"); ok {
		port, err := strconv.Atoi(grpcPort)
		if err != nil {
			return nil, fmt.Errorf("failed to parse grpc port %s", grpcPort) 
		}
		cfg.GrpcPort = port 
	}

	if jsonConfig, ok := os.LookupEnv("CONFIG"); ok {
		cfg.JSONConfig = jsonConfig
	}

	return e.NextParser.Parse(cfg)
}

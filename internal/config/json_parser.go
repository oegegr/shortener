package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// JSONConfigParser парсит конфигурацию из JSON файла
type JSONConfigParser struct {
	NextParser
}

// парсит конфигурацию из JSON файла
func (j *JSONConfigParser) Parse(cfg *Config) (*Config, error) {
	jsonConfig, err := j.tryLoadConfigFromJSONFile()
	if err != nil {
		return nil, err
	}

	if jsonConfig != nil {
		j.mergeConfig(cfg, jsonConfig)
	}

	return j.NextParser.Parse(cfg)
}

// пробуем загрузит JSON из файла
func (j *JSONConfigParser) tryLoadConfigFromJSONFile() (*Config, error) {
	configFile := j.getConfigFilePath()
	if configFile == "" {
		return nil, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var jsonConfig Config
	if err := json.Unmarshal(data, &jsonConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &jsonConfig, nil
}

func (j *JSONConfigParser) getConfigFilePath() string {
	var configFileShort string
	var configFileLong string

	fs := flag.NewFlagSet("config-finder", flag.ContinueOnError)
	fs.StringVar(&configFileShort, "c", "", "Config file path (short)")
	fs.StringVar(&configFileLong, "config", "", "Config file path (long)")

	_ = fs.Parse(os.Args[1:])

	if configFileLong != "" {
		return configFileLong
	}
	if configFileShort != "" {
		return configFileShort
	}

	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		return envConfig
	}

	return ""
}

// мержим конфигурацию приложениея с JSON конфигурацией
func (j *JSONConfigParser) mergeConfig(main *Config, json *Config) {
	if json.ServerAddress != "" {
		main.ServerAddress = json.ServerAddress
	}
	if json.BaseURL != "" {
		main.BaseURL = json.BaseURL
	}
	if json.FileStoragePath != "" {
		main.FileStoragePath = json.FileStoragePath
	}
	if json.DBConnectionString != "" {
		main.DBConnectionString = json.DBConnectionString
	}
	if json.LogLevel != "" {
		main.LogLevel = json.LogLevel
	}
	if json.JWTSecret != "" {
		main.JWTSecret = json.JWTSecret
	}
	if json.AuditFile != "" {
		main.AuditFile = json.AuditFile
	}
	if json.AuditURL != "" {
		main.AuditURL = json.AuditURL
	}
	if json.TLSCertFile != "" {
		main.TLSCertFile = json.TLSCertFile
	}
	if json.TLSKeyFile != "" {
		main.TLSKeyFile = json.TLSKeyFile
	}
	main.EnableHTTPS = json.EnableHTTPS
	if json.ShortURLLength > 0 {
		main.ShortURLLength = json.ShortURLLength
	}
}

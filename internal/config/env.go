package config

import "os"

func NewEnvConfig() *Config {
	cfg := &Config{
		ServerAddress: os.Getenv("SERVER_ADDRESS"),
		BaseURL: os.Getenv("BASE_URL"),
	}
	return cfg
}
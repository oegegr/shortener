package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress  string
	BaseURL        string
	ShortURLLength int
}

func NewConfig() *Config {
	cfg := &Config{}
	cfg.ServerAddress = os.Getenv("SERVER_ADDRESS")
	cfg.BaseURL = os.Getenv("BASE_URL")

	if cfg.ServerAddress == "" {
		flag.StringVar(&cfg.ServerAddress, "a", "127.0.0.1:8080", "address to startup server")
	}

	if cfg.BaseURL == "" {
		flag.StringVar(&cfg.BaseURL, "b", "http://127.0.0.1", "domain to use for shrten urls")
	}

	flag.IntVar(&cfg.ShortURLLength, "c", 8, "length of generated short url")
	flag.Parse()

	return cfg
}

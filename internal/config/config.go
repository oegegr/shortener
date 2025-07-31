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

	flag.StringVar(&cfg.ServerAddress, "a", "127.0.0.1:8080", "address to startup server")
	flag.StringVar(&cfg.BaseURL, "b", "http://127.0.0.1:8080", "domain to use for shrten urls")
	flag.IntVar(&cfg.ShortURLLength, "c", 8, "length of generated short url")
	flag.Parse()

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}

	return cfg
}
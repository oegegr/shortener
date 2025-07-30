package config

import (
	"flag"
)


func NewFlagConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", "127.0.0.1:8080", "address to startup server")
	flag.StringVar(&cfg.BaseURL, "b", "http://127.0.0.1:8080", "domain to use for shrten urls")
	flag.IntVar(&cfg.ShortURLLength, "c", 8, "length of generated short url")
	flag.Parse()

	return cfg
}

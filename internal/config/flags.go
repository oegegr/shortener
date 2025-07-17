package config

import "flag"

type Config struct {
	RunAddr string
	ShortURLDomain string
	ShortURLLength int
}

func NewConfig() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.RunAddr, "a", ":8080", "address to startup server")
	flag.StringVar(&cfg.ShortURLDomain, "b", "http://localhost:8080", "domain to use for shrten urls")
	flag.IntVar(&cfg.ShortURLLength, "c", 8, "length of generated short url")
	flag.Parse()
	return cfg
}

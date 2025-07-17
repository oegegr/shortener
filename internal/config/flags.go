package config

import "flag"

type Config struct {
	RunAddr string
	ShortURLDomain string
	ShortURLLength int
}

var AppConfig Config

func ParseFlags() {
	flag.StringVar(&AppConfig.RunAddr, "a", ":8080", "address to startup server")
	flag.StringVar(&AppConfig.ShortURLDomain, "b", "http://localhost:8080", "domain to use for shrten urls")
	flag.IntVar(&AppConfig.ShortURLLength, "c", 8, "length of generated short url")
	flag.Parse()
}
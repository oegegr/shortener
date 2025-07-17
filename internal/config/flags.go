package config

import "flag"

type Config struct {
	RunAddr string
	ShortUrlDomain string
	ShortUrlLength int
}

var AppConfig Config

func ParseFlags() {
	flag.StringVar(&AppConfig.RunAddr, "a", ":8080", "address to startup server")
	flag.StringVar(&AppConfig.ShortUrlDomain, "b", "http://localhost:8080", "domain to use for shrten urls")
	flag.IntVar(&AppConfig.ShortUrlLength, "c", 8, "length of generated short url")
	flag.Parse()
}
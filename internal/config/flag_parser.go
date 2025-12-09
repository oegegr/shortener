package config

import "flag"

// FlagsConfigParser парсит конфигурацию из флагов командной строки
type FlagsConfigParser struct {
	NextParser
}

// парсит конфигурацию из флагов командной строки
func (f *FlagsConfigParser) Parse(cfg *Config) (*Config, error) {
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "address to startup server")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "domain to use for short urls")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file path to save storage")
	flag.StringVar(&cfg.DBConnectionString, "d", cfg.DBConnectionString, "database connection string")
	flag.StringVar(&cfg.LogLevel, "l", cfg.LogLevel, "log level")
	flag.StringVar(&cfg.JWTSecret, "jwtkey", cfg.JWTSecret, "jwt secret key")
	flag.StringVar(&cfg.AuditFile, "audit-file", cfg.AuditFile, "file to keep audit logs")
	flag.StringVar(&cfg.AuditURL, "audit-url", cfg.AuditURL, "URL to pass audit logs")
	flag.IntVar(&cfg.ShortURLLength, "short-len", cfg.ShortURLLength, "length of generated short url")
	flag.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "Enable HTTPS")
	flag.StringVar(&cfg.TLSCertFile, "tlscert", cfg.TLSCertFile, "TLS certificate file")
	flag.StringVar(&cfg.TLSKeyFile, "tlskey", cfg.TLSKeyFile, "TLS key file")
	flag.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "Trusted subnet which has permission to process api")
	flag.IntVar(&cfg.GrpcPort, "grpcport", cfg.GrpcPort, "If present configure grpc server")

	var configFileShort string
	var configFileLong string

	flag.StringVar(&configFileShort, "c", "", "Config file path (short)")
	flag.StringVar(&configFileLong, "config", "", "Config file path (long)")

	if configFileLong != "" {
		cfg.JSONConfig = configFileLong
	}
	if configFileShort != "" {
		cfg.JSONConfig = configFileShort
	}

	flag.Parse()

	return f.NextParser.Parse(cfg)
}

package config 

type Config struct {
	ServerAddress  string
	BaseURL        string
	ShortURLLength int
}



func NewConfig() *Config {
	cfg := &Config{}
	envConfig := NewEnvConfig()
	flagConfig := NewFlagConfig()


	if envConfig.ServerAddress == "" {
		cfg.ServerAddress = flagConfig.ServerAddress
	} else {
		cfg.ServerAddress = envConfig.ServerAddress
	}
	if envConfig.BaseURL == "" {
		cfg.BaseURL = flagConfig.BaseURL
	
	} else {
		cfg.BaseURL = flagConfig.BaseURL
	}

	cfg.ShortURLLength = flagConfig.ShortURLLength

	return cfg
}
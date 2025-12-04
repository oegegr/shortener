// Package config содержит реализацию конфигурации приложения.
package config

// Config представляет структуру конфигурации приложения.
type Config struct {
	// ServerAddress представляет адрес сервера.
	ServerAddress string `json:"server_address,omitempty"`
	// BaseURL представляет базовый URL-адрес для сокращенных URL-адресов.
	BaseURL string `json:"base_url,omitempty"`
	// ShortURLLength представляет длину сокращенного URL-адреса.
	ShortURLLength int `json:"short_url_length,omitempty"`
	// FileStoragePath представляет путь к файлу для хранения данных.
	FileStoragePath string `json:"file_storage_path,omitempty"`
	// DBConnectionString представляет строку подключения к базе данных.
	DBConnectionString string `json:"db_connection_string,omitempty"`
	// LogLevel представляет уровень логирования.
	LogLevel string `json:"log_level,omitempty"`
	// JWTSecret представляет секретный ключ для JWT-токенов.
	JWTSecret string `json:"jwt_secret,omitempty"`
	// AuditFile представляет файл для хранения аудит-логов.
	AuditFile string `json:"audit_file,omitempty"`
	// AuditURL представляет URL-адрес для отправки аудит-логов.
	AuditURL string `json:"audit_url,omitempty"`
	// Включения HTTPS
	EnableHTTPS bool `json:"enable_https,omitempty"`
	// Путь TLS к сертификату
	TLSCertFile string `json:"tls_cert_file,omitempty"`
	// Путь TLS к ключу
	TLSKeyFile string `json:"tls_key_file,omitempty"`
	// Доверенная сеть которой можно отвечать
	TrustedSubnet string `json:"trusted_subnet,omitempty"`
	// Путь JSON конфигу
	JSONConfig string
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		ServerAddress:   "127.0.0.1:8080",
		BaseURL:         "http://127.0.0.1:8080",
		ShortURLLength:  8,
		FileStoragePath: "",
		LogLevel:        "DEBUG",
		JWTSecret:       "jwt-secret-key",
		AuditFile:       "",
		AuditURL:        "",
		EnableHTTPS:     false,
		TLSCertFile:     "cert.pem",
		TLSKeyFile:      "key.pem",
		TrustedSubnet:   "",
	}
}

// ConfigParser интерфейс для парсеров конфигурации
type ConfigParser interface {
	Parse(cfg *Config) (*Config, error)
}

// NextParser базовая структура для парсеров
type NextParser struct {
	next ConfigParser
}

func (b *NextParser) Parse(cfg *Config) (*Config, error) {
	if b.next != nil {
		return b.next.Parse(cfg)
	}
	return cfg, nil
}

// NewConfig возвращает новый экземпляр конфигурации приложения.
// Эта функция парсит флаги командной строки и переменные окружения для инициализации конфигурации.
func NewConfig() (*Config, error) {
	cfg := DefaultConfig()

	// Создаем цепочку вручную
	envParser := &EnvConfigParser{}
	flagsParser := &FlagsConfigParser{NextParser: NextParser{next: envParser}}
	jsonParser := &JSONConfigParser{NextParser: NextParser{next: flagsParser}}

	return jsonParser.Parse(cfg)
}

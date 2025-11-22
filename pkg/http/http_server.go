package https

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Server интерфейс для запуска HTTP сервера
type Server interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// HTTPServer реализация для HTTP
type httpServer struct {
	server *http.Server
}

func (s *httpServer) Start(ctx context.Context) error {
	return s.server.ListenAndServe()
}

func (s *httpServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// HTTPSServer реализация для HTTPS
type httpsServer struct {
	server      *http.Server
	tlsCertFile string
	tlsKeyFile  string
}

func (s *httpsServer) Start(ctx context.Context) error {
	return s.server.ListenAndServeTLS(s.tlsCertFile, s.tlsKeyFile)
}

func (s *httpsServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// ServerBuilder конфигурирует HTTP сервер с поддержкой TLS
type ServerBuilder struct {
	serverAddress string
	enableHTTPS   bool
	tlsCertFile   string
	tlsKeyFile    string
	readTimeout   time.Duration
	writeTimeout  time.Duration
	idleTimeout   time.Duration
}

// NewServerBuilder создает новый билдер сервера
func NewServerBuilder(serverAddress string) *ServerBuilder {
	return &ServerBuilder{
		serverAddress: serverAddress,
		readTimeout:   10 * time.Second,
		writeTimeout:  10 * time.Second,
		idleTimeout:   60 * time.Second,
	}
}

// WithHTTPS включает поддержку HTTPS с указанием файлов сертификатов
func (b *ServerBuilder) WithHTTPS(certFile, keyFile string) *ServerBuilder {
	b.enableHTTPS = true
	b.tlsCertFile = certFile
	b.tlsKeyFile = keyFile
	return b
}

// WithTimeouts устанавливает таймауты для сервера
func (b *ServerBuilder) WithTimeouts(read, write, idle time.Duration) *ServerBuilder {
	b.readTimeout = read
	b.writeTimeout = write
	b.idleTimeout = idle
	return b
}

// Build создает и настраивает HTTP сервер
func (b *ServerBuilder) Build(handler http.Handler) (Server, error) {
	server := &http.Server{
		Addr:         b.serverAddress,
		Handler:      handler,
		ReadTimeout:  b.readTimeout,
		WriteTimeout: b.writeTimeout,
		IdleTimeout:  b.idleTimeout,
	}

	if b.enableHTTPS {
		if err := b.configureTLS(server); err != nil {
			return nil, err
		}
		return &httpsServer{
			server:      server,
			tlsCertFile: b.tlsCertFile,
			tlsKeyFile:  b.tlsKeyFile,
		}, nil
	}

	return &httpServer{server: server}, nil
}

// configureTLS настраивает TLS для сервера
func (b *ServerBuilder) configureTLS(server *http.Server) error {
	if err := b.checkTLSFiles(); err != nil {
		return err
	}

	server.TLSConfig = &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP256, tls.X25519},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	return nil
}

// checkTLSFiles проверяет существование файлов сертификатов
func (b *ServerBuilder) checkTLSFiles() error {
	if _, err := os.Stat(b.tlsCertFile); os.IsNotExist(err) {
		return fmt.Errorf("TLS certificate file not found: %s", b.tlsCertFile)
	}
	if _, err := os.Stat(b.tlsKeyFile); os.IsNotExist(err) {
		return fmt.Errorf("TLS key file not found: %s", b.tlsKeyFile)
	}
	return nil
}

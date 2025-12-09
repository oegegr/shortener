package http

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
)

// GrpcServerBuilder конфигурирует gRPC сервер с поддержкой TLS
type GrpcServerBuilder struct {
	serverAddress string
	grpcServer    *grpc.Server
	enableHTTPS   bool
	tlsCertFile   string
	tlsKeyFile    string
}

// WithHTTPS включает поддержку HTTPS с указанием файлов сертификатов
func (b *GrpcServerBuilder) WithHTTPS(certFile, keyFile string) *GrpcServerBuilder {
	b.enableHTTPS = true
	b.tlsCertFile = certFile
	b.tlsKeyFile = keyFile
	return b
}

// NewServerBuilder создает новый билдер сервера
func NewGrpcServerBuilder(serverAddress string, grpcServer *grpc.Server) *GrpcServerBuilder {
	return &GrpcServerBuilder{
		serverAddress: serverAddress,
		grpcServer:    grpcServer,
	}
}

// Build создает и настраивает gRPC сервер
func (b *GrpcServerBuilder) Build() (Server, error) {
	return &grpcServer{
		server:  b.grpcServer,
		address: b.serverAddress,
	}, nil
}

// configureTLS настраивает TLS для сервера
func (b *GrpcServerBuilder) configureTLS() error {
	if err := b.checkTLSFiles(); err != nil {
		return err
	}
	return nil
}

// checkTLSFiles проверяет существование файлов сертификатов
func (b *GrpcServerBuilder) checkTLSFiles() error {
	if _, err := os.Stat(b.tlsCertFile); os.IsNotExist(err) {
		return fmt.Errorf("TLS certificate file not found: %s", b.tlsCertFile)
	}
	if _, err := os.Stat(b.tlsKeyFile); os.IsNotExist(err) {
		return fmt.Errorf("TLS key file not found: %s", b.tlsKeyFile)
	}
	return nil
}

// loadTLSCertificate загружает сертификат TLS
func (b *GrpcServerBuilder) loadTLSCertificate() tls.Certificate {
	cert, err := tls.LoadX509KeyPair(b.tlsCertFile, b.tlsKeyFile)
	if err != nil {
		log.Fatal(err)
	}
	return cert
}

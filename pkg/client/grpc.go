package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/oegegr/shortener/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// GRPCClient представляет клиент для взаимодействия с gRPC сервером
type GRPCClient struct {
	conn    *grpc.ClientConn
	client  api.ShortenerServiceClient
	address string
	token   string
}

// Config конфигурация клиента
type Config struct {
	Address  string
	UseTLS   bool
	CertFile string
	Insecure bool // Для self-signed сертификатов
	Timeout  time.Duration
}

// NewGRPCClient создает новый gRPC клиент
func NewGRPCClient(cfg Config) (*GRPCClient, error) {
	var opts []grpc.DialOption
	
	// Настраиваем безопасность
	if cfg.UseTLS {
		var creds credentials.TransportCredentials
		if cfg.Insecure {
			creds = credentials.NewTLS(&tls.Config{
				InsecureSkipVerify: true,
			})
		} else if cfg.CertFile != "" {
			var err error
			creds, err = credentials.NewClientTLSFromFile(cfg.CertFile, "")
			if err != nil {
				return nil, fmt.Errorf("failed to load TLS cert: %w", err)
			}
		} else {
			creds = credentials.NewTLS(&tls.Config{})
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	
	// Настраиваем таймауты
	if cfg.Timeout > 0 {
		opts = append(opts, grpc.WithTimeout(cfg.Timeout))
	}
	
	// Устанавливаем соединение
	conn, err := grpc.Dial(cfg.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
	}
	
	return &GRPCClient{
		conn:    conn,
		client:  api.NewShortenerServiceClient(conn),
		address: cfg.Address,
	}, nil
}

// SetToken устанавливает токен авторизации
func (c *GRPCClient) SetToken(token string) {
	c.token = token
}

// createContext создает контекст с метаданными для авторизации
func (c *GRPCClient) createContext(ctx context.Context) context.Context {
	if c.token != "" {
		md := metadata.Pairs("authorization", c.token)
		return metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

// ShortenURL сокращает URL
func (c *GRPCClient) ShortenURL(ctx context.Context, url string) (string, error) {
	req := &api.URLShortenRequest{Url: url}
	
	resp, err := c.client.ShortenURL(c.createContext(ctx), req)
	if err != nil {
		return "", fmt.Errorf("failed to shorten URL: %w", err)
	}
	
	return resp.Result, nil
}

// ExpandURL возвращает оригинальный URL
func (c *GRPCClient) ExpandURL(ctx context.Context, id string) (string, error) {
	req := &api.URLExpandRequest{Id: id}
	
	resp, err := c.client.ExpandURL(c.createContext(ctx), req)
	if err != nil {
		return "", fmt.Errorf("failed to expand URL: %w", err)
	}
	
	return resp.Result, nil
}

// ListUserURLs возвращает список URL пользователя
func (c *GRPCClient) ListUserURLs(ctx context.Context) ([]*api.URLData, error) {
	req := &api.Empty{}
	
	resp, err := c.client.ListUserURLs(c.createContext(ctx), req)
	if err != nil {
		return nil, fmt.Errorf("failed to list user URLs: %w", err)
	}
	
	return resp.Url, nil
}

// Close закрывает соединение
func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Ping проверяет доступность сервера
func (c *GRPCClient) Ping(ctx context.Context) error {
	// Для простоты используем ListUserURLs как ping
	_, err := c.ListUserURLs(ctx)
	return err
}
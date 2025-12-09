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
	userID  string
}

// Config конфигурация клиента
type Config struct {
	Address  string
	UseTLS   bool
	CertFile string
	Insecure bool
	Timeout  time.Duration
	Token    string
	UserID   string
}

// NewGRPCClient создает новый gRPC клиент с использованием grpc.NewClient
func NewGRPCClient(cfg Config) (*GRPCClient, error) {
	// Для grpc.NewClient используем grpc.WithTransportCredentials
	var transportCreds credentials.TransportCredentials
	
	if cfg.UseTLS {
		if cfg.Insecure {
			transportCreds = credentials.NewTLS(&tls.Config{
				InsecureSkipVerify: true,
			})
		} else if cfg.CertFile != "" {
			var err error
			transportCreds, err = credentials.NewClientTLSFromFile(cfg.CertFile, "")
			if err != nil {
				return nil, fmt.Errorf("failed to load TLS cert: %w", err)
			}
		} else {
			transportCreds = credentials.NewTLS(&tls.Config{})
		}
	} else {
		transportCreds = insecure.NewCredentials()
	}
	
	// Создаем соединение с помощью grpc.NewClient
	conn, err := grpc.NewClient(
		cfg.Address,
		grpc.WithTransportCredentials(transportCreds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	
	return &GRPCClient{
		conn:    conn,
		client:  api.NewShortenerServiceClient(conn),
		address: cfg.Address,
		token:   cfg.Token,
		userID:  cfg.UserID,
	}, nil
}

// SetToken устанавливает токен авторизации
func (c *GRPCClient) SetToken(token string) {
	c.token = token
}

// SetUserID устанавливает идентификатор пользователя
func (c *GRPCClient) SetUserID(userID string) {
	c.userID = userID
}

// GetUserID возвращает текущий userID
func (c *GRPCClient) GetUserID() string {
	return c.userID
}

// createContext создает контекст с метаданными для авторизации
func (c *GRPCClient) createContext(ctx context.Context) context.Context {
	md := metadata.New(nil)
	
	// Добавляем токен если есть
	if c.token != "" {
		md.Set("authorization", c.token)
	}
	
	// Добавляем userID если есть
	if c.userID != "" {
		md.Set("x-user-id", c.userID)
	}
	
	return metadata.NewOutgoingContext(ctx, md)
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
	// Пытаемся вызвать простой метод
	req := &api.Empty{}
	_, err := c.client.ListUserURLs(c.createContext(ctx), req)
	return err
}
package http

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

// grpcServer реализация для gRPC
type grpcServer struct {
	server   *grpc.Server
	address  string
	listener net.Listener
}

// NewGrpcServer создает новый gRPC сервер
func NewGrpcServer(address string, server *grpc.Server) *grpcServer {
	return &grpcServer{
		server:  server,
		address: address,
	}
}

// Start запускает gRPC сервер
func (s *grpcServer) Start(ctx context.Context) error {
	// Создаем listener при старте
	var err error
	s.listener, err = net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.address, err)
	}

	// Канал для ошибок сервера
	serverErr := make(chan error, 1)
	
	// Запускаем сервер в горутине
	go func() {
		if err := s.server.Serve(s.listener); err != nil {
			serverErr <- fmt.Errorf("gRPC server error: %w", err)
		}
		close(serverErr)
	}()

	// Ждем либо ошибку, либо завершение контекста
	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		// Graceful shutdown при отмене контекста
		s.server.GracefulStop()
		return nil
	}
}

// Stop останавливает gRPC сервер
func (s *grpcServer) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	// Graceful shutdown с таймаутом
	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		// Успешная остановка
	case <-ctx.Done():
		// Таймаут - принудительная остановка
		s.server.Stop()
	}

	// Закрываем listener если он был создан
	if s.listener != nil {
		s.listener.Close()
	}

	return nil
}
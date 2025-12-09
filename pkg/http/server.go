package http

import "context"

// Server интерфейс для запуска HTTP сервера
type Server interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Package db содержит реализацию подключения к базе данных PostgreSQL.
package db

import (
	"database/sql"
	"errors"

	"github.com/oegegr/shortener/internal/config"
	"go.uber.org/zap"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewDB возвращает новый экземпляр подключения к базе данных PostgreSQL.
// Эта функция принимает конфигурацию приложения и логгер, и возвращает подключение к базе данных и ошибку.
func NewDB(c config.Config, logger *zap.SugaredLogger) (*sql.DB, error) {
	db, err := sql.Open("pgx", c.DBConnectionString)

	if err != nil {
		logger.Fatal("failed to create db connection %w", err.Error())
		return nil, err
	}

	m, err := migrate.New("file://migrations", c.DBConnectionString)
	if err != nil {
		logger.Fatal("failed to configure db migrations %w", err.Error())
		return nil, err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Fatal("failed to apply db migrations %w", err.Error())
		return nil, err
	}

	return db, nil
}

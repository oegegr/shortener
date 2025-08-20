package db

import (
	"database/sql"
	"errors"

	"github.com/oegegr/shortener/internal/config"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func NewDB(c config.Config, logger *zap.SugaredLogger) (*sql.DB, error){
	if c.DBConnectionString == "" {
		return nil, nil
	}
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
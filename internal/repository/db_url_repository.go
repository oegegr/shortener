package repository

import (
	"context"
	"database/sql"

	"github.com/oegegr/shortener/internal/model"
	"go.uber.org/zap"
)

type DBURLRepository struct {
	db *sql.DB
	logger zap.SugaredLogger
}

func NewDBURLRepository(db *sql.DB, logger zap.SugaredLogger) *DBURLRepository {
	return &DBURLRepository{
		db: db,
		logger: logger,
	}
}

func (r *DBURLRepository) CreateURL(ctx context.Context, urlItem []model.URLItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err 
	}
	stmt, err := tx.Prepare("INSERT INTO url (url, short_id) VALUES ($1, $2)")
	if err != nil {
		r.logger.Errorf("sql request validation error: %v", err)
		return err
	}
	defer stmt.Close()

	for _, item := range urlItem {
		_, err = stmt.Exec(item.URL, item.ShortID)
		if err != nil {
			r.logger.Errorf("sql request execution error: %v", err)
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
 
}

func (r *DBURLRepository) FindURLByID(ctx context.Context, id string) (*model.URLItem, error) {
	stmt, err := r.db.Prepare("SELECT url, short_id FROM url WHERE short_id = $1")
	if err != nil {
		r.logger.Errorf("sql validation error: %v", err)
		return nil, err
	}
	defer stmt.Close()

	var urlItem model.URLItem
	err = stmt.QueryRow(id).Scan(&urlItem.URL, &urlItem.ShortID)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debugf("url not found %s", id)
			return nil, nil
		}
		r.logger.Errorf("sql execution error: %v", err)
		return nil, err
	}

	return &urlItem, nil
}

func (r *DBURLRepository) Exists(ctx context.Context, id string) bool {
	stmt, err := r.db.Prepare("SELECT 1 FROM url WHERE id = $1")
	if err != nil {
		r.logger.Errorf("sql validation error: %v", err)
		return false
	}
	defer stmt.Close()

	var exists bool
	err = stmt.QueryRow(id).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debugf("url not found %s", id)
			return false
		}
		r.logger.Errorf("sql execution error: %v", err)
		return false
	}

	return exists
}

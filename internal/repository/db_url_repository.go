package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/oegegr/shortener/internal/model"
	"go.uber.org/zap"
)

type DBURLRepository struct {
	db *sql.DB
	logger zap.SugaredLogger
}

func NewDBURLRepository(db *sql.DB, logger zap.SugaredLogger) (*DBURLRepository, error) {
	return &DBURLRepository{
		db: db,
		logger: logger,
	}, nil
}

func (r *DBURLRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *DBURLRepository) CreateURL(ctx context.Context, urlItem []model.URLItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err 
	}
	stmt, err := tx.Prepare("INSERT INTO url (url, short_id, user_id) VALUES ($1, $2, $3)")
	if err != nil {
		r.logger.Errorf("sql request validation error: %v", err)
		return err
	}
	defer stmt.Close()

	for _, item := range urlItem {

		_, err = stmt.Exec(item.URL, item.ShortID, item.UserID)

		if err != nil {
			if strings.Contains(err.Error(), "23505") {
					return ErrRepoURLAlreadyExists 
				}

			r.logger.Errorf("sql request execution error: %v", err)
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

 

func (r *DBURLRepository) FindURLByURL(ctx context.Context, url string) (*model.URLItem, error) {
	stmt, err := r.db.Prepare("SELECT url, short_id , user_id FROM url WHERE url = $1")
	if err != nil {
		r.logger.Errorf("sql validation error: %v", err)
		return nil, err
	}
	defer stmt.Close()

	var urlItem model.URLItem
	err = stmt.QueryRow(url).Scan(&urlItem.URL, &urlItem.ShortID, &urlItem.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debugf("url not found %s", url)
			return nil, ErrRepoNotFound 
		}
		r.logger.Errorf("sql execution error: %v", err)
		return nil, err
	}

	return &urlItem, nil
}

func (r *DBURLRepository) FindURLByUser(ctx context.Context, userID string) ([]model.URLItem, error) {
	stmt, err := r.db.Prepare("SELECT url, short_id, user_id FROM url WHERE user_id = $1")
	if err != nil {
		r.logger.Errorf("sql validation error: %v", err)
		return nil, err
	}
	defer stmt.Close()

	var items []model.URLItem
	rows, err := stmt.Query(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debugf("user not found %s", )
			return nil, ErrRepoNotFound 
		}
		r.logger.Errorf("sql execution error: %v", err)
		return nil, err
	}
	for rows.Next() {
		var item model.URLItem
		rows.Scan(&item.URL, &item.ShortID, &item.UserID)
		items = append(items, item)
	}

	return items, nil
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

package model

import (
	"errors"
	"time"
)

var (
	ErrInvalidURL = errors.New("invalid URL format")
	ErrEmptyCode  = errors.New("short code cannot be empty")
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type UserURLResponse []UserURL

type UserURL struct {
	ShortURL string `json:"short_url"`
	URL      string `json:"original_url"`
}

type ShortenBatchDeleteRequest []string

type ShortenBatchRequest []BatchRequest

type BatchRequest struct {
	URL           string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

type ShortenBatchResponse []BatchResponse

type BatchResponse struct {
	Result        string `json:"short_url"`
	CorrelationID string `json:"correlation_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type URLItem struct {
	ShortID   string `json:"short_id"`
	URL       string `json:"original_url"`
	UserID    string `json:"user_id"`
	IsDeleted bool   `json:"id_deleted"`
}

type LogAuditItem struct {
	TS     int64     `json:"ts"`
	Action LogAction `json:"action"`
	UserID *string   `json:"user_id,omitempty"`
	URL    string    `json:"url"`
}

func NewLogAuditItem(url string, userID string, action LogAction) *LogAuditItem {
	var user *string
	if userID != "" {
		user = &userID
	}
	return &LogAuditItem{
		TS:     time.Now().Unix(),
		Action: action,
		UserID: user,
		URL:    url,
	}
}

type LogAction string

const (
	LogActionShorten LogAction = "shorten"
	LogActionFollow  LogAction = "follow"
)

func NewURLItem(url string, id string, userID string, isDeleted bool) *URLItem {
	return &URLItem{
		URL:       url,
		ShortID:   id,
		UserID:    userID,
		IsDeleted: isDeleted,
	}
}

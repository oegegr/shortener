// Package model содержит структуры данных и константы для работы с URL-адресами.
package model

import (
	"errors"
	"time"
)

// ErrInvalidURL представляет ошибку, которая возникает при неверном формате URL-адреса.
var ErrInvalidURL = errors.New("invalid URL format")

// ErrEmptyCode представляет ошибку, которая возникает при пустом коде сокращения URL-адреса.
var ErrEmptyCode = errors.New("short code cannot be empty")

// ShortenRequest представляет запрос на сокращение URL-адреса.
type ShortenRequest struct {
	// URL представляет URL-адрес, который необходимо сократить.
	URL string `json:"url"`
}

// ShortenResponse представляет ответ на запрос на сокращение URL-адреса.
type ShortenResponse struct {
	// Result представляет сокращенный URL-адрес.
	Result string `json:"result"`
}

// UserURLResponse представляет список URL-адресов пользователя.
type UserURLResponse []UserURL

// UserURL представляет URL-адрес пользователя.
type UserURL struct {
	// ShortURL представляет сокращенный URL-адрес.
	ShortURL string `json:"short_url"`
	// URL представляет оригинальный URL-адрес.
	URL string `json:"original_url"`
}

// ShortenBatchDeleteRequest представляет запрос на удаление нескольких URL-адресов.
type ShortenBatchDeleteRequest []string

// ShortenBatchRequest представляет запрос на сокращение нескольких URL-адресов.
type ShortenBatchRequest []BatchRequest

// BatchRequest представляет запрос на сокращение URL-адреса с корреляционным идентификатором.
type BatchRequest struct {
	// URL представляет URL-адрес, который необходимо сократить.
	URL string `json:"original_url"`
	// CorrelationID представляет корреляционный идентификатор запроса.
	CorrelationID string `json:"correlation_id"`
}

// ShortenBatchResponse представляет ответ на запрос на сокращение нескольких URL-адресов.
type ShortenBatchResponse []BatchResponse

// BatchResponse представляет ответ на запрос на сокращение URL-адреса с корреляционным идентификатором.
type BatchResponse struct {
	// Result представляет сокращенный URL-адрес.
	Result string `json:"short_url"`
	// CorrelationID представляет корреляционный идентификатор запроса.
	CorrelationID string `json:"correlation_id"`
}

// ErrorResponse представляет ответ с ошибкой.
type ErrorResponse struct {
	// Error представляет текст ошибки.
	Error string `json:"error"`
}

// URLItem представляет элемент URL-адреса.
type URLItem struct {
	// ShortID представляет сокращенный идентификатор URL-адреса.
	ShortID string `json:"short_id"`
	// URL представляет оригинальный URL-адрес.
	URL string `json:"original_url"`
	// UserID представляет идентификатор пользователя.
	UserID string `json:"user_id"`
	// IsDeleted представляет флаг, указывающий, удален ли URL-адрес.
	IsDeleted bool `json:"id_deleted"`
}

// NewURLItem возвращает новый элемент URL-адреса.
// Эта функция принимает URL-адрес, сокращенный идентификатор, идентификатор пользователя и флаг удаления.
func NewURLItem(url string, id string, userID string, isDeleted bool) *URLItem {
	return &URLItem{
		// ...
	}
}

// LogAuditItem представляет элемент аудита логов.
type LogAuditItem struct {
	// TS представляет метку времени аудита.
	TS int64 `json:"ts"`
	// Action представляет действие аудита.
	Action LogAction `json:"action"`
	// UserID представляет идентификатор пользователя, если доступен.
	UserID *string `json:"user_id,omitempty"`
	// URL представляет URL-адрес, связанный с аудитом.
	URL string `json:"url"`
}

// NewLogAuditItem возвращает новый элемент аудита логов.
// Эта функция принимает URL-адрес, идентификатор пользователя и действие аудита.
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

// LogAction представляет тип действия аудита.
type LogAction string

// LogActionShorten представляет действие сокращения URL-адреса.
const LogActionShorten LogAction = "shorten"

// LogActionFollow представляет действие перехода по URL-адресу.
const LogActionFollow LogAction = "follow"

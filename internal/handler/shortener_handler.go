// Package handler содержит обработчики HTTP-запросов для работы с URL-адресами.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	app_error "github.com/oegegr/shortener/internal/error"
	"github.com/oegegr/shortener/internal/model"
	"github.com/oegegr/shortener/internal/repository"
	"github.com/oegegr/shortener/internal/service"
)

const (
	shortenFailure = "failed to get short url"
)

// UserIDProvider провайдер предоставляет доступ к индентификатору пользователя 
type UserIDProvider interface {
	// Get возвращает идентификатор пользователя из контекста запроса.
	Get(ctx context.Context) (string, error)
}

// ShortenerHandler обрабатывает HTTP-запросы для работы с URL-адресами.
type ShortenerHandler struct {
	// URLService предоставляет сервис для работы с URL-адресами.
	URLService service.URLShortener
	// userIDProvider предоставляет провайдер для получения идентификатора пользователя.
	userIDProvider UserIDProvider
	// logAudit предоставляет менеджер для аудита логов.
	logAudit service.LogAuditManager
}

// NewShortenerHandler возвращает новый экземпляр ShortenerHandler.
func NewShortenerHandler(
	service service.URLShortener,
	provider UserIDProvider,
	logAudit service.LogAuditManager,
) ShortenerHandler {
	return ShortenerHandler{
		URLService:     service,
		userIDProvider: provider,
		logAudit:       logAudit,
	}
}

// RedirectToOriginalURL обрабатывает HTTP-запрос на перенаправление на оригинальный URL-адрес.
func (app *ShortenerHandler) RedirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, _ := app.userIDProvider.Get(ctx)

	shortURL := chi.URLParam(r, "short_url")
	if shortURL == "" {
		http.Error(w, "missing short url at params", http.StatusBadRequest)
		return
	}

	originalURL, err := app.URLService.GetOriginalURL(ctx, shortURL)
	if err != nil {

		if errors.Is(err, app_error.ErrServiceURLGone) {
			http.Error(w, err.Error(), http.StatusGone)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	app.logAudit.NotifyAllAuditors(ctx, *model.NewLogAuditItem(originalURL, userID, model.LogActionFollow))
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

// ShortenURL обрабатывает HTTP-запрос на сокращение URL-адреса.
func (app *ShortenerHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	url := string(body)

	if err != nil {
		http.Error(w, "missing post body", http.StatusBadRequest)
		return
	}

	err = validateURL(url)
	if err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	userID, err := app.userIDProvider.Get(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL, err := app.URLService.GetShortURL(ctx, url, userID)
	if err != nil {

		if errors.Is(err, repository.ErrRepoURLAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(shortURL))
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	app.logAudit.NotifyAllAuditors(ctx, *model.NewLogAuditItem(url, userID, model.LogActionShorten))
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

// APIUserURL обрабатывает HTTP-запрос на получение URL-адресов пользователя.
func (app *ShortenerHandler) APIUserURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := app.userIDProvider.Get(ctx)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	urlItems, err := app.URLService.GetUserURL(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(urlItems) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := model.UserURLResponse(urlItems)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// APIUserBatchDeleteURL обрабатывает HTTP-запрос на удаление URL-адресов пользователя в пакетном режиме.
func (app *ShortenerHandler) APIUserBatchDeleteURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req model.ShortenBatchDeleteRequest

	if r.Header.Get("Content-type") != "application/json" {
		http.Error(w, "wrong content-type", http.StatusBadRequest)
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "failed to deserialize body", http.StatusBadRequest)
		return
	}

	userID, err := app.userIDProvider.Get(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = app.URLService.DeleteUserURL(ctx, userID, req)
	if err != nil {
		if errors.Is(err, service.ErrDeleteQueueIsFull) {
			http.Error(w, err.Error(), http.StatusTooManyRequests)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// APIShortenBatchURL обрабатывает HTTP-запрос на сокращение URL-адресов в пакетном режиме.
func (app *ShortenerHandler) APIShortenBatchURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req model.ShortenBatchRequest

	if r.Header.Get("Content-type") != "application/json" {
		http.Error(w, "wrong content-type", http.StatusBadRequest)
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "failed to deserialize body", http.StatusBadRequest)
		return
	}

	urls := []string{}
	for _, item := range req {
		err := validateURL(item.URL)
		if err != nil {
			http.Error(w, "invalid url", http.StatusBadRequest)
			return
		}
		urls = append(urls, item.URL)
	}

	userID, err := app.userIDProvider.Get(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortURLs, err := app.URLService.GetShortURLBatch(ctx, urls, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := model.ShortenBatchResponse{}
	for idx, shortURL := range shortURLs {
		item := model.BatchResponse{
			CorrelationID: req[idx].CorrelationID,
			Result:        shortURL,
		}
		resp = append(resp, item)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// APIShortenURL обрабатывает HTTP-запрос на сокращение URL-адреса.
func (app *ShortenerHandler) APIShortenURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req model.ShortenRequest

	var err error

	if r.Header.Get("Content-type") != "application/json" {
		http.Error(w, "wrong content-type", http.StatusBadRequest)
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "failed to deserialize body", http.StatusBadRequest)
		return
	}

	err = validateURL(req.URL)
	if err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	userID, err := app.userIDProvider.Get(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL, err := app.URLService.GetShortURL(ctx, req.URL, userID)
	if err != nil {

		if errors.Is(err, repository.ErrRepoURLAlreadyExists) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(model.ShortenResponse{Result: shortURL})
			return
		}

		http.Error(w, shortenFailure, http.StatusBadRequest)
		return
	}

	app.logAudit.NotifyAllAuditors(ctx, *model.NewLogAuditItem(req.URL, userID, model.LogActionShorten))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.ShortenResponse{Result: shortURL})

}

// validateURL проверяет корректность URL-адреса.
func validateURL(originalURL string) error {
	if _, err := url.ParseRequestURI(originalURL); err != nil {
		return err
	}
	return nil
}

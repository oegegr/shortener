// Package handler содержит обработчики HTTP-запросов.
package handler

import (
	"net/http"

	"github.com/oegegr/shortener/internal/repository"
)

// PingHandler обрабатывает запросы на эндпоинт /ping.
type PingHandler struct {
	// repo хранит ссылку на репозиторий URL-адресов.
	repo repository.URLRepository
}

// NewPingHandler возвращает новый экземпляр PingHandler.
func NewPingHandler(repo repository.URLRepository) PingHandler {
	return PingHandler{repo: repo}
}

// Ping обрабатывает HTTP-запрос на эндпоинт /ping и проверяет подключение к базе данных.
func (p *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := p.repo.Ping(ctx)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

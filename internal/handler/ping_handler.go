package handler

import (
	"net/http"

	"github.com/oegegr/shortener/internal/repository"
)

type PingHandler struct {
	repo repository.URLRepository
}

func NewPingHandler(repo repository.URLRepository) PingHandler {
	return PingHandler{repo: repo}
}

func (p *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := p.repo.Ping(ctx)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

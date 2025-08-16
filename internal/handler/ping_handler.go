package handler

import (
	"net/http"

	"database/sql"
)

type PingHandler struct {
	db *sql.DB
}

func NewPingHandler(db *sql.DB) PingHandler {
	return PingHandler{db: db}
}

func (p *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	if p.db == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := p.db.PingContext(ctx); err != nil {
		http.Error(w, "failed to connect to database", http.StatusInternalServerError)
		return 
    }

	w.WriteHeader(http.StatusOK)
}
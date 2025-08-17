package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"errors"

	"github.com/go-chi/chi/v5"
	"github.com/oegegr/shortener/internal/model"
	"github.com/oegegr/shortener/internal/repository"
	"github.com/oegegr/shortener/internal/service"
)

type ShortenerHandler struct {
	URLService service.URLShortener
}

func NewShortenerHandler(service service.URLShortener) ShortenerHandler {
	return ShortenerHandler{URLService: service}
}

func (app *ShortenerHandler) RedirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shortURL := chi.URLParam(r, "short_url")
	if shortURL == "" {
		http.Error(w, "missing short url at params", http.StatusBadRequest)
		return
	}

	originalURL, err := app.URLService.GetOriginalURL(ctx, shortURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

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

	shortURL, err := app.URLService.GetShortURL(ctx, url)
	if err != nil {

		if errors.Is(err, repository.ErrRepoURLAlreadyExists) {
			// http.Error(w, shortURL, http.StatusConflict)
			w.WriteHeader(http.StatusConflict)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(shortURL))
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

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

	shortURLs, err := app.URLService.GetShortURLBatch(ctx, urls)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := model.ShortenBatchResponse{}
	for idx, shortURL := range shortURLs {
		item := model.BatchResponse{
			CorrelationID: req[idx].CorrelationID,
			Result: shortURL,
		}
		resp = append(resp, item)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

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

	shortURL, err := app.URLService.GetShortURL(ctx, req.URL)
	if err != nil {

		if errors.Is(err, repository.ErrRepoURLAlreadyExists) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(model.ShortenResponse{Result: shortURL})
			return
		}

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.ShortenResponse{Result: shortURL})

}

func validateURL(originalURL string) error {
	if _, err := url.ParseRequestURI(originalURL); err != nil {
		return err
	}
	return nil
}

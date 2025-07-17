package handler

import (
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/oegegr/shortener/internal/service"
)

type ShortnerHandler struct {
	UrlService service.URLShortner
}

func NewShortnerHandler(service service.URLShortner) ShortnerHandler {
	return ShortnerHandler{UrlService: service}
}

func (app *ShortnerHandler) RedirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "wrong method", http.StatusBadRequest)
		return
	}
	shortURL := chi.URLParam(r, "short_url")
	if shortURL == "" {
		http.Error(w, "missing short url at params", http.StatusBadRequest)
		return
	}

	originalURL, err := app.UrlService.GetOriginalURL(shortURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func (app *ShortnerHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "wrong http method", http.StatusBadRequest)
		return
	}

	if contentType := r.Header.Get("Content-type"); contentType != "text/plain" {
		http.Error(w, "wrong content type", http.StatusBadRequest)
		return
	}

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

	shortURL, err := app.UrlService.GetShortURL(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func validateURL(originalURL string) error {
	if _, err := url.ParseRequestURI(originalURL); err != nil {
		return err
	}
	return nil
}

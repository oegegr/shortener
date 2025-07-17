package handler

import (
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/oegegr/shortener/internal/service"
)

type ShortenerHandler struct {
	URLService service.URLShortener
}

func NewShortenerHandler(service service.URLShortener) ShortenerHandler {
	return ShortenerHandler{URLService: service}
}

func (app *ShortenerHandler) RedirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "short_url")
	if shortURL == "" {
		http.Error(w, "missing short url at params", http.StatusBadRequest)
		return
	}

	originalURL, err := app.URLService.GetOriginalURL(shortURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func (app *ShortenerHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
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

	shortURL, err := app.URLService.GetShortURL(url)
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

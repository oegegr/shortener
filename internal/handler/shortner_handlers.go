package handler

import (
	"io"
	"net/http"
	"net/url"

	"github.com/oegegr/shortener/internal/service"
)

type ShortnerHandler struct {
	urlService service.ShortenURLService
}

func NewShortnerHandler(service service.ShortenURLService) ShortnerHandler {
	return ShortnerHandler{urlService: service}
}

func (app *ShortnerHandler) RedirectToOriginalURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	shortURL := r.PathValue("short_url")
	if shortURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL, err := app.urlService.GetOriginalURL(shortURL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func (app *ShortnerHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	url := string(body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = validateURL(url)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := app.urlService.GetShortURL(url)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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

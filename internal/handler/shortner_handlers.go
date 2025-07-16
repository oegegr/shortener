package handler

import (
	"io"
	"net/http"

	"github.com/oegegr/shortener/internal/service"
)

type ShortnerHandler struct {
	urlService service.ShortenUrlService
}

func NewShortnerHandler(service service.ShortenUrlService) (ShortnerHandler) {
	return ShortnerHandler{urlService: service} 
}

func (app *ShortnerHandler) RedirectToOriginalUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	shortUrl:= r.PathValue("short_url")
	if shortUrl == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalUrl, err := app.urlService.GetOriginalUrl(shortUrl)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, originalUrl, http.StatusTemporaryRedirect)
}

func (app *ShortnerHandler) ShortenUrl(w http.ResponseWriter, r *http.Request) {
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
	shortUrl, err := app.urlService.GetShortUrl(url)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(shortUrl))
}

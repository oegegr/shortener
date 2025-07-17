package handler_test

import (
	"net/http/httptest"
	"net/http"
	"testing"
	"errors"
	"context"
	"strings"
	"io"

	"github.com/stretchr/testify/assert"
	"github.com/go-chi/chi/v5"
	"github.com/oegegr/shortener/internal/handler"
	"github.com/oegegr/shortener/internal/service"
)

func TestRedirectToOriginalUrl(t *testing.T) {
	service := new(service.MockURLService)
	app := handler.NewShortnerHandler(service)

	t.Run("Invalid Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		w := httptest.NewRecorder()
		app.RedirectToOriginalURL(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("Empty Short URL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		app.RedirectToOriginalURL(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("Service Error", func(t *testing.T) {
		service.On("GetOriginalURL", "abc").Return("", errors.New("error"))
		
		req := httptest.NewRequest(http.MethodGet, "/abc", nil)
		w := httptest.NewRecorder()
		app.RedirectToOriginalURL(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("Valid Redirect", func(t *testing.T) {
		service.On("GetOriginalURL", "xyz").Return("https://google.com", nil)

		
		req := httptest.NewRequest(http.MethodGet, "/xyz", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("short_url", "xyz")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		app.RedirectToOriginalURL(w, req)
		
		resp := w.Result()
		assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
		assert.Equal(t, "https://google.com", resp.Header.Get("Location"))
	})
}

func TestShortenUrl(t *testing.T) {
	service := new(service.MockURLService)
	app := handler.NewShortnerHandler(service)

	t.Run("Valid Shortening", func(t *testing.T) {
		body := strings.NewReader("https://google.com")
		req := httptest.NewRequest(http.MethodPost, "/", body)
		// req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()
		
		service.On("GetShortURL", "https://google.com").Return("abc123", nil)
		app.ShortenURL(w, req)
		
		resp := w.Result()
		bodyBytes, _ := io.ReadAll(resp.Body)
		
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
		assert.Equal(t, "abc123", string(bodyBytes))
	})

	t.Run("Invalid Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		app.ShortenURL(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("Invalid Content-type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Content-type", "application/json")
		w := httptest.NewRecorder()
		app.ShortenURL(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("Read Body Error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", errReader(0))
		w := httptest.NewRecorder()
		app.ShortenURL(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("Service Error", func(t *testing.T) {
		body := strings.NewReader("https://google.com")
		req := httptest.NewRequest(http.MethodPost, "/", body)
		w := httptest.NewRecorder()
		
		service.On("GetShortURL", "https://google.com").Return("", errors.New("error"))
		app.ShortenURL(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})
}

type errReader int
func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}
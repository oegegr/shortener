package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/oegegr/shortener/internal/handler"
	"github.com/oegegr/shortener/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserIDProvider struct {
	mock.Mock
}

func (m *MockUserIDProvider) Get(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func TestRedirectToOriginalUrl(t *testing.T) {
	service := new(service.MockURLService)
	userIdProvider := new(handler.UserIDProvider)
	app := handler.NewShortenerHandler(service, *userIdProvider)

	t.Run("Invalid Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		w := httptest.NewRecorder()
		app.RedirectToOriginalURL(w, req)
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("Empty Short URL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		app.RedirectToOriginalURL(w, req)
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("Service Error", func(t *testing.T) {
		service.On("GetOriginalURL", "abc").Return("", errors.New("error"))

		req := httptest.NewRequest(http.MethodGet, "/abc", nil)
		w := httptest.NewRecorder()
		app.RedirectToOriginalURL(w, req)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("Valid Redirect", func(t *testing.T) {

		service.On("GetOriginalURL", mock.Anything, "xyz").Return("https://google.com", nil)

		req := httptest.NewRequest(http.MethodGet, "/xyz", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("short_url", "xyz")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		app.RedirectToOriginalURL(w, req)

		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
		assert.Equal(t, "https://google.com", res.Header.Get("Location"))
	})
}

func TestShortenUrl(t *testing.T) {
	service := new(service.MockURLService)
	userIdProvider := new(MockUserIDProvider)
	app := handler.NewShortenerHandler(service, userIdProvider)
	
	userIdProvider.On("Get", mock.Anything).Return("user", nil)

	t.Run("Valid Shortening", func(t *testing.T) {
		body := strings.NewReader("https://google.com")
		req := httptest.NewRequest(http.MethodPost, "/", body)
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		service.On("GetShortURL", mock.Anything, "https://google.com", "user").Return("abc123", nil).Once()
		app.ShortenURL(w, req)

		res := w.Result()
		bodyBytes, _ := io.ReadAll(res.Body)
		defer res.Body.Close()

		assert.Equal(t, http.StatusCreated, res.StatusCode)
		assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))
		assert.Equal(t, "abc123", string(bodyBytes))
	})

	t.Run("Invalid Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		app.ShortenURL(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("Invalid Content-type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Content-type", "application/json")
		w := httptest.NewRecorder()
		app.ShortenURL(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("Service Error", func(t *testing.T) {
		body := strings.NewReader("https://google.com")
		req := httptest.NewRequest(http.MethodPost, "/", body)
		w := httptest.NewRecorder()

		service.On("GetShortURL", mock.Anything, "https://google.com", "user").Return("", errors.New("error")).Once()
		app.ShortenURL(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	})

	t.Run("Read Body Error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", errReader(0))
		w := httptest.NewRecorder()
		app.ShortenURL(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

func TestApiShortenUrl(t *testing.T) {
	service := new(service.MockURLService)
	userIdProvider := new(MockUserIDProvider)
	app := handler.NewShortenerHandler(service, userIdProvider)

	userIdProvider.On("Get", mock.Anything).Return("user", nil)

	t.Run("Valid Shortening", func(t *testing.T) {
		reqBody := map[string]string{"url": "https://google.com"}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		service.On("GetShortURL", mock.Anything, "https://google.com", "user").Return("abc123", nil).Once()
		app.APIShortenURL(w, req)

		res := w.Result()
		bodyBytes, _ := io.ReadAll(res.Body)
		defer res.Body.Close()

		assert.Equal(t, http.StatusCreated, res.StatusCode)
		assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
		assert.JSONEq(t, `{"result": "abc123"}`, string(bodyBytes))
	})

	t.Run("Invalid Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/shorten", nil)
		w := httptest.NewRecorder()
		app.APIShortenURL(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("Invalid Content-type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/shorten", nil)
		req.Header.Set("Content-type", "text/plain")
		w := httptest.NewRecorder()
		app.APIShortenURL(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("Service Error", func(t *testing.T) {
		body := strings.NewReader("https://google.com")
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
		w := httptest.NewRecorder()

		service.On("GetShortURL", "https://google.com").Return("", errors.New("error")).Once()
		app.APIShortenURL(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("Read Body Error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", errReader(0))
		w := httptest.NewRecorder()
		app.APIShortenURL(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

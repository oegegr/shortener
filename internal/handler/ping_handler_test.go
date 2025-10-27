// Package handler содержит примеры работы с обработчиками HTTP-запросов.
package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oegegr/shortener/internal/repository"
	"github.com/stretchr/testify/mock"
)

func TestPingHandler_Ping(t *testing.T) {
	// Создание фейкового репозитория
	repo := new(repository.MockURLRepository)

	repo.On("Ping", mock.Anything).Return(nil).Once()

	// Создание обработчика
	handler := NewPingHandler(repo)

	// Создание запроса
	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Создание записи запроса
	w := httptest.NewRecorder()

	// Обработка запроса
	handler.Ping(w, req)

	// Проверка статуса ответа
	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, but got %d", http.StatusOK, w.Code)
	}
}

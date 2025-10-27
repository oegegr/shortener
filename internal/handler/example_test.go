// Package handler содержит примеры работы с обработчиками HTTP-запросов.
package handler

import (
	"net/http"
	"net/http/httptest"

	"github.com/oegegr/shortener/internal/repository"
	"github.com/stretchr/testify/mock"
)

func ExamplePingHandler() {
	// Создание фейкового репозитория
	repo := new(repository.MockURLRepository)

	repo.On("Ping", mock.Anything).Return(nil).Once()

	// Создание обработчика
	handler := NewPingHandler(repo)

	// Создание запроса
	req, _ := http.NewRequest("GET", "/ping", nil)

	// Создание записи запроса
	w := httptest.NewRecorder()

	// Обработка запроса
	handler.Ping(w, req)
}

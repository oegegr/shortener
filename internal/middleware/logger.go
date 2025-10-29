// Package middleware содержит middleware-функции для логирования HTTP-запросов.
package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseRecorder представляет запись ответа, которая отслеживает статус и размер ответа.
type responseRecorder struct {
	http.ResponseWriter
	// status представляет статус ответа.
	status int
	// size представляет размер ответа в байтах.
	size int
}

// WriteHeader отправляет заголовок ответа и устанавливает статус ответа.
func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// Write записывает данные в ответ и обновляет размер ответа.
func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}

// ZapLogger возвращает middleware-функцию для логирования HTTP-запросов с помощью Zap.
// Эта функция принимает экземпляр логгера Zap и возвращает middleware-функцию.
func ZapLogger(sugar zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			method := r.Method
			uri := r.RequestURI
			startTime := time.Now()
			recorderedWriter := &responseRecorder{ResponseWriter: w}
			nextHandler.ServeHTTP(recorderedWriter, r)
			duration := time.Since(startTime)
			sugar.Infoln(
				"request",
				"uri", uri,
				"method", method,
				"duration", duration,
				"status code", recorderedWriter.status,
				"response size", recorderedWriter.size,
			)
		})
	}
}

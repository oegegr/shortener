package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}

func ZapLogger(sugar zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			uri := r.RequestURI
			method := r.Method
			recorderedWriter := &responseRecorder{ResponseWriter: w}
			nextHandler.ServeHTTP(recorderedWriter, r)
			duration := time.Since(start)
			sugar.Infoln(
				"request",
				"uri", uri,
				"method", method,
				"duration", duration,
				"status code", recorderedWriter.status,
				"response size", recorderedWriter.size,
			)

		}
		return http.HandlerFunc(logFn)
	}

}

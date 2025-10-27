// Package middleware содержит middleware-функции для сжатия HTTP-ответов.
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// GzipMiddleware возвращает middleware-функцию для сжатия HTTP-ответов с помощью gzip.
// Эта функция принимает список типов контента, которые должны быть сжаты.
func GzipMiddleware(compressedTypes []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		gzipFunc := func(w http.ResponseWriter, r *http.Request) {
			ow := w

			acceptEncoding := r.Header.Get("Accept-Encoding")
			shouldCompress := strings.Contains(acceptEncoding, "gzip")
			if shouldCompress {
				cw := newCompressWriter(w, compressedTypes)
				ow = cw
				defer cw.Close()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer cr.Close()
			}

			next.ServeHTTP(ow, r)
		}
		return http.HandlerFunc(gzipFunc)
	}
}

// compressWriter представляет writer, который сжимает данные с помощью gzip.
type compressWriter struct {
	// responseWriter представляет оригинальный writer.
	responseWriter http.ResponseWriter
	// zipWriter представляет writer, который сжимает данные с помощью gzip.
	zipWriter *gzip.Writer
	// compressedTypes представляет список типов контента, которые должны быть сжаты.
	compressedTypes []string
	// wroteHeader представляет флаг, который указывает, был ли отправлен заголовок ответа.
	wroteHeader bool
	// shouldCompress представляет флаг, который указывает, следует ли сжимать данные.
	shouldCompress bool
}

// newCompressWriter возвращает новый экземпляр compressWriter.
// Эта функция принимает оригинальный writer и список типов контента, которые должны быть сжаты.
func newCompressWriter(w http.ResponseWriter, ct []string) *compressWriter {
	return &compressWriter{
		responseWriter:  w,
		compressedTypes: ct,
	}
}

// Header возвращает заголовки ответа.
func (c *compressWriter) Header() http.Header {
	return c.responseWriter.Header()
}

// Write записывает данные в ответ.
// Если сжатие включено, данные будут сжаты с помощью gzip.
func (c *compressWriter) Write(p []byte) (int, error) {
	if !c.wroteHeader {
		c.WriteHeader(http.StatusOK)
	}

	if c.shouldCompress {
		if c.zipWriter == nil {
			c.zipWriter = gzip.NewWriter(c.responseWriter)
		}
		return c.zipWriter.Write(p)
	}
	return c.responseWriter.Write(p)
}

// WriteHeader отправляет заголовок ответа.
// Если сжатие включено, в заголовке будет указано, что содержимое сжато с помощью gzip.
func (c *compressWriter) WriteHeader(statusCode int) {
	if c.wroteHeader {
		return
	}
	c.wroteHeader = true

	if statusCode < 300 {
		c.shouldCompress = c.needCompression()
		if c.shouldCompress {
			c.responseWriter.Header().Set("Content-Encoding", "gzip")
		}
	}

	c.responseWriter.WriteHeader(statusCode)
}

// Close закрывает writer и освобождает ресурсы.
func (c *compressWriter) Close() error {
	if c.zipWriter != nil {
		return c.zipWriter.Close()
	}
	return nil
}

// needCompression проверяет, следует ли сжимать данные на основе типа контента.
func (c *compressWriter) needCompression() bool {
	contentType := c.responseWriter.Header().Get("Content-Type")
	for _, t := range c.compressedTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}
	return false
}

// compressReader представляет reader, который распаковывает сжатые данные с помощью gzip.
type compressReader struct {
	// r представляет оригинальный reader.
	r io.ReadCloser
	// zr представляет reader, который распаковывает сжатые данные с помощью gzip.
	zr *gzip.Reader
}

// newCompressReader возвращает новый экземпляр compressReader.
// Эта функция принимает оригинальный reader и возвращает reader, который распаковывает сжатые данные.
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read читает данные из запроса.
// Если данные сжаты, они будут распакованы с помощью gzip.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close закрывает reader и освобождает ресурсы.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

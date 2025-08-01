package middleware

import (
	"net/http"
	"compress/gzip"
    "io"
	"strings"
)

func GzipMiddleware(compressedTypes []string) func(http.Handler) http.Handler {
return func(next http.Handler) http.Handler {
    gzipFunc := func(w http.ResponseWriter, r *http.Request) {
        // по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
        // который будем передавать следующей функции
        ow := w

        // проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
        acceptEncoding := r.Header.Get("Accept-Encoding")

        shouldCompressed := strings.Contains(acceptEncoding, "gzip")
        if shouldCompressed {
            // оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
            cw := newCompressWriter(w, compressedTypes)
            // меняем оригинальный http.ResponseWriter на новый
            ow = cw
            // не забываем отправить клиенту все сжатые данные после завершения middleware
            defer cw.Close()

        }

        // проверяем, что клиент отправил серверу сжатые данные в формате gzip
        contentEncoding := r.Header.Get("Content-Encoding")
        sendsGzip := strings.Contains(contentEncoding, "gzip")
        if sendsGzip {
            // оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
            cr, err := newCompressReader(r.Body)
            if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                return
            }
            // меняем тело запроса на новое
            r.Body = cr
            defer cr.Close()
        }

        // передаём управление хендлеру
        next.ServeHTTP(ow, r)
    }
	return http.HandlerFunc(gzipFunc)
}
}

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
    responseWriter  http.ResponseWriter
    zipWriter *gzip.Writer
	compressedTypes []string
}

func newCompressWriter(w http.ResponseWriter, ct []string) *compressWriter {
    return &compressWriter{
        responseWriter:  w,
        zipWriter: gzip.NewWriter(w),
		compressedTypes: ct,
    }
}

func (c *compressWriter) Header() http.Header {
    return c.responseWriter.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if c.shouldCompress() {
		c.responseWriter.Header().Set("Content-Encoding", "gzip")
		return c.zipWriter.Write(p)
	} 
	return c.responseWriter.Write(p)

}

func (c *compressWriter) WriteHeader(statusCode int) {
    if statusCode < 300 && c.shouldCompress(){
        c.responseWriter.Header().Set("Content-Encoding", "gzip")
    }
    c.responseWriter.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
    return c.zipWriter.Close()
}

func (c *compressWriter) shouldCompress() bool {
		contentType := c.responseWriter.Header().Get("Content-Type")
		for  _, t := range c.compressedTypes {
			if strings.Contains(contentType, t) {
				return true 
			}
		}
		return false
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
    r  io.ReadCloser
    zr *gzip.Reader
}

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

func (c compressReader) Read(p []byte) (n int, err error) {
    return c.zr.Read(p)
}

func (c *compressReader) Close() error {
    if err := c.r.Close(); err != nil {
        return err
    }
    return c.zr.Close()
} 
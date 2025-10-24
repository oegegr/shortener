package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

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

type compressWriter struct {
	responseWriter  http.ResponseWriter
	zipWriter       *gzip.Writer
	compressedTypes []string
	wroteHeader     bool
	shouldCompress  bool
}

func newCompressWriter(w http.ResponseWriter, ct []string) *compressWriter {
	return &compressWriter{
		responseWriter:  w,
		compressedTypes: ct,
	}
}

func (c *compressWriter) Header() http.Header {
	return c.responseWriter.Header()
}

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

func (c *compressWriter) Close() error {
	if c.zipWriter != nil {
		return c.zipWriter.Close()
	}
	return nil
}

func (c *compressWriter) needCompression() bool {
	contentType := c.responseWriter.Header().Get("Content-Type")
	for _, t := range c.compressedTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}
	return false
}

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

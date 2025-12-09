package http

import (
	"context"
	"net/http"
)

// HTTPServer реализация для HTTP
type httpServer struct {
	server *http.Server
}

func (s *httpServer) Start(ctx context.Context) error {
	return s.server.ListenAndServe()
}

func (s *httpServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// HTTPSServer реализация для HTTPS
type httpsServer struct {
	server      *http.Server
	tlsCertFile string
	tlsKeyFile  string
}

func (s *httpsServer) Start(ctx context.Context) error {
	return s.server.ListenAndServeTLS(s.tlsCertFile, s.tlsKeyFile)
}

func (s *httpsServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

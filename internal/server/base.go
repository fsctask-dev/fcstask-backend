package server

import (
	"context"
	"errors"
	"net/http"
)

type HTTPServer interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type httpServer struct {
	httpServer *http.Server
}

func NewHTTPServer(addr string, handler http.Handler) HTTPServer {
	return &httpServer{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
}

func (s *httpServer) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		_ = s.httpServer.Shutdown(context.Background())
	}()

	err := s.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *httpServer) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

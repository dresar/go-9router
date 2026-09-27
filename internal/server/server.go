package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dresar/go-9router/internal/logging"
)

type Server struct {
	http *http.Server
	addr string
}

func New(addr string, handler http.Handler, readHeaderTimeout, idleTimeout time.Duration) *Server {
	return &Server{
		addr: addr,
		http: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: readHeaderTimeout,
			IdleTimeout:       idleTimeout,
		},
	}
}

func (s *Server) Start() error {
	logging.Info("SERVER", fmt.Sprintf("listening on http://%s", s.addr))
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

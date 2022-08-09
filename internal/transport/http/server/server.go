package server

import (
	"golang.org/x/net/context"
	"net/http"
	"service-account/internal/config"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(config *config.Config, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:    config.HTTP.Host,
			Handler: handler,
		},
	}
}

func (server *Server) Run() error {
	return server.httpServer.ListenAndServe()
}

func (server *Server) Stop(context context.Context) error {
	return server.httpServer.Shutdown(context)
}

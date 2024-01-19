package server

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Server interface {
	ListenAndServe() error
}

type WebServer struct {
	host   string
	router chi.Router
}

// ListenAndServe implements Server.
func (ws *WebServer) ListenAndServe() error {
	return http.ListenAndServe(ws.host, ws.router)
}

func NewWebServer(host string, router chi.Router) *WebServer {
	return &WebServer{host: host, router: router}
}

var _ Server = (*WebServer)(nil)

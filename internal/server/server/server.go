package server

import (
	"net/http"

	"github.com/shadyziedan/metrica/internal/server/handlers"
)

type Server interface {
	ListenAndServe() error
}

type WebServer struct {
	host       string
	repository handlers.MetricsRepository
}

// ListenAndServe implements Server.
func (ws *WebServer) ListenAndServe() error {
	router := handlers.NewRouter(ws.repository)
	return http.ListenAndServe(ws.host, router)
}

func NewWebServer(host string, repository handlers.MetricsRepository) *WebServer {
	return &WebServer{host: host, repository: repository}
}

var _ Server = (*WebServer)(nil)

package server

import (
	"net/http"

	"github.com/shadyziedan/metrica/internal/server/handlers"
	"github.com/shadyziedan/metrica/internal/server/storage"
)

type Server interface {
	ListenAndServe() error
}

type WebServer struct {
	host       string
	repository storage.MetricsRepository
}

// ListenAndServe implements Server.
func (ws *WebServer) ListenAndServe() error {
	router := handlers.NewRouter(ws.repository)
	return http.ListenAndServe(ws.host, router)
}

func NewWebServer(host string, repository storage.MetricsRepository) *WebServer {
	return &WebServer{host: host, repository: repository}
}

var _ Server = (*WebServer)(nil)

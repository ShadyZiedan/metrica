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
	host string
	repository storage.MetricsRepository 
}

// ListenAndServe implements Server.
func (ws *WebServer) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.Handle(`/update/counter/`, handlers.UpdateMetricHandler(ws.repository))
	mux.Handle(`/update/gauge/`, handlers.UpdateMetricHandler(ws.repository))
	mux.Handle(`/update/`, http.HandlerFunc(handlers.UnknownMetricHandler))
	return http.ListenAndServe(ws.host, mux)
}

func NewWebServer(host string, repository storage.MetricsRepository) *WebServer {
	return &WebServer{host: host, repository: repository}
}

var _ Server = (*WebServer)(nil)

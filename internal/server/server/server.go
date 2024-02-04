package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

//type Server interface {
//	ListenAndServe(ctx context.Context) error
//}

type WebServer struct {
	http.Server
}

func (ws *WebServer) ListenAndServe(ctx context.Context) error {
	go func() {
		err := ws.Server.ListenAndServe()
		if err != nil {
			return
		}
	}()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return ws.Server.Shutdown(shutdownCtx)
}

func NewWebServer(host string, router chi.Router) *WebServer {
	return &WebServer{
		Server: http.Server{
			Addr:    host,
			Handler: router,
		},
	}
}

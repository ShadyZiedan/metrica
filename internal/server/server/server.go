package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type WebServer struct {
	http.Server
}

func (ws *WebServer) ListenAndServe(ctx context.Context) error {
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		err := ws.Server.Shutdown(shutdownCtx)
		if err != nil {
			panic(err)
		}
	}()
	err := ws.Server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func NewWebServer(host string, router chi.Router) *WebServer {
	return &WebServer{
		Server: http.Server{
			Addr:    host,
			Handler: router,
		},
	}
}

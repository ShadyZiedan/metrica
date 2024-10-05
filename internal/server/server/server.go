// Package server provides the functionality for creating and managing a web server.
// It includes the WebServer struct, which embeds the http.Server type, and provides methods for starting and stopping the server.
package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// WebServer represents a web server.
// It embeds the http.Server type and provides methods for starting and stopping the server.
type WebServer struct {
	http.Server
}

// ListenAndServe starts the web server and listens for incoming HTTP requests.
// It accepts a context as a parameter, which can be used to cancel the server gracefully.
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

// NewWebServer creates a new instance of the WebServer struct.
// It accepts a host string and a chi.Router as parameters, and returns a pointer to the new WebServer instance.
func NewWebServer(host string, router chi.Router) *WebServer {
	return &WebServer{
		Server: http.Server{
			Addr:    host,
			Handler: router,
		},
	}
}

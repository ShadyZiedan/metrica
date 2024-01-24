package main

import (
	"github.com/shadyziedan/metrica/internal/server/config"
	"github.com/shadyziedan/metrica/internal/server/handlers"
	"github.com/shadyziedan/metrica/internal/server/logger"
	"github.com/shadyziedan/metrica/internal/server/middleware"
	"github.com/shadyziedan/metrica/internal/server/server"
	"github.com/shadyziedan/metrica/internal/server/storage"
)

func main() {
	cnf := config.ParseConfig()
	memStorage := storage.NewMemStorage()
	err := logger.Initialize("info")
	if err != nil {
		panic(err)
	}
	router := handlers.NewRouter(memStorage, middleware.RequestLogger)
	srv := server.NewWebServer(
		cnf.Address,
		router,
	)
	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

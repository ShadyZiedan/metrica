package main

import (
	"github.com/shadyziedan/metrica/internal/server/config"
	"github.com/shadyziedan/metrica/internal/server/handlers"
	"github.com/shadyziedan/metrica/internal/server/server"
	"github.com/shadyziedan/metrica/internal/server/storage"
)

func main() {
	cnf := config.ParseConfig()
	memStorage := storage.NewMemStorage()
	router := handlers.NewRouter(memStorage)
	srv := server.NewWebServer(
		cnf.Address,
		router,
	)
	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

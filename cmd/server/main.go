package main

import (
	"github.com/shadyziedan/metrica/internal/server/config"
	"github.com/shadyziedan/metrica/internal/server/server"
	"github.com/shadyziedan/metrica/internal/server/storage"
)

func main() {
	config := config.ParseConfig()

	storage := storage.NewMemStorage()
	server := server.NewWebServer(config.Address, storage)
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

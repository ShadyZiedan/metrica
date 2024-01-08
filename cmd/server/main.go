package main

import (
	"github.com/shadyziedan/metrica/internal/server/server"
	"github.com/shadyziedan/metrica/internal/server/storage"
)

func main() {
	storage := storage.NewMemStorage()
	server := server.NewWebServer(":8080", storage)
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

package main

import (
	"flag"

	"github.com/shadyziedan/metrica/internal/server/server"
	"github.com/shadyziedan/metrica/internal/server/storage"
)

func main() {
	address := flag.String("a", "localhost:8080", "адрес эндпоинта HTTP-сервера")
	flag.Parse()
	
	storage := storage.NewMemStorage()
	server := server.NewWebServer(*address, storage)
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

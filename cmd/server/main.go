package main

import (
	"flag"

	"github.com/caarlos0/env"
	"github.com/shadyziedan/metrica/internal/server/server"
	"github.com/shadyziedan/metrica/internal/server/storage"
)

type Config struct {
	Address string `env:"ADDRESS" envDefault:"localhost:8080"`
}

func parseConfig() Config {
	var config Config

	err := env.Parse(&config)
	if err != nil {
		panic(err)
	}

	flag.StringVar(&config.Address, "a", config.Address, "адрес эндпоинта HTTP-сервера")

	flag.Parse()
	return config
}

func main() {
	config := parseConfig()

	storage := storage.NewMemStorage()
	server := server.NewWebServer(config.Address, storage)
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

package config

import (
	"flag"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Address string `env:"ADDRESS" envDefault:"localhost:8080"`
}

func ParseConfig() Config {
	var config Config

	err := env.Parse(&config)
	if err != nil {
		panic(err)
	}

	flag.StringVar(&config.Address, "a", config.Address, "адрес эндпоинта HTTP-сервера")

	flag.Parse()
	return config
}

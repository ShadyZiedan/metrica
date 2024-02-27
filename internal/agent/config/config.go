package config

import (
	"flag"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Address        string `env:"ADDRESS" envDefault:"localhost:8080"`
	ReportInterval int    `env:"REPORT_INTERVAL" envDefault:"10"`
	PollInterval   int    `env:"POLL_INTERVAL" envDefault:"2"`
	Key            string `env:"KEY"`
}

func ParseConfig() Config {
	var config Config

	err := env.Parse(&config)
	if err != nil {
		panic(err)
	}

	flag.StringVar(&config.Address, "a", config.Address, "адрес эндпоинта HTTP-сервера")
	flag.IntVar(&config.ReportInterval, "r", config.ReportInterval, "частота отправки метрик на сервер")
	flag.IntVar(&config.PollInterval, "p", config.PollInterval, "частота опроса метрик из пакета runtime")
	flag.StringVar(&config.Key, "k", config.Key, "Ключ")

	flag.Parse()
	return config
}

package config

import (
	"flag"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
}

func ParseConfig() Config {
	var config Config

	flag.StringVar(&config.Address, "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")
	flag.IntVar(&config.ReportInterval, "r", 10, "частота отправки метрик на сервер")
	flag.IntVar(&config.PollInterval, "p", 2, "частота опроса метрик из пакета runtime")
	flag.StringVar(&config.Key, "k", "", "Ключ")

	flag.Parse()
	err := env.Parse(&config)
	if err != nil {
		panic(err)
	}

	return config
}

// Package config provides a way to parse and access configuration settings for the agent.
package config

import (
	"flag"

	"github.com/caarlos0/env/v10"
)

// Config represents the configuration settings for the agent.
type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

// ParseConfig parses the command-line flags and environment variables to create a new Config instance.
func ParseConfig() Config {
	var config Config

	flag.StringVar(&config.Address, "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")
	flag.IntVar(&config.ReportInterval, "r", 10, "частота отправки метрик на сервер")
	flag.IntVar(&config.PollInterval, "p", 2, "частота опроса метрик из пакета runtime")
	flag.StringVar(&config.Key, "k", "", "Ключ")
	flag.IntVar(&config.RateLimit, "l", 1, "Rate limit")

	flag.Parse()
	err := env.Parse(&config)
	if err != nil {
		panic(err)
	}

	return config
}

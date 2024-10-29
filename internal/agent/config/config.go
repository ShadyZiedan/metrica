// Package config provides a way to parse and access configuration settings for the agent.
package config

import (
	"flag"

	"github.com/caarlos0/env/v10"
)

// Config represents the configuration settings for the agent.
type Config struct {
	// Address is web server host address
	Address string `env:"ADDRESS"`
	// ReportInterval is interval in seconds of reporting metrics
	ReportInterval int `env:"REPORT_INTERVAL"`
	// PollInterval is interval or frequency in seconds of gathering metrics from runtime package
	PollInterval int `env:"POLL_INTERVAL"`
	// Key is a secret key for hashing data
	Key string `env:"KEY"`
	// RateLimit is a rate limiter of metrics being sent
	RateLimit int `env:"RATE_LIMIT"`
	// CryptoKey is a path to public key to encrypt data sent to the server
	CryptoKey string `env:"CRYPTO_KEY"`
}

// ParseConfig parses the command-line flags and environment variables to create a new Config instance.
func ParseConfig() Config {
	var config Config

	flag.StringVar(&config.Address, "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")
	flag.IntVar(&config.ReportInterval, "r", 10, "частота отправки метрик на сервер")
	flag.IntVar(&config.PollInterval, "p", 2, "частота опроса метрик из пакета runtime")
	flag.StringVar(&config.Key, "k", "", "Ключ")
	flag.IntVar(&config.RateLimit, "l", 1, "Rate limit")
	flag.StringVar(&config.CryptoKey, "crypto-key", "", "путь до файла с публичным ключом")

	flag.Parse()
	err := env.Parse(&config)
	if err != nil {
		panic(err)
	}

	return config
}

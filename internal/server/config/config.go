// Package config provides a way to parse and access configuration settings for the web server.
package config

import (
	"flag"

	"github.com/caarlos0/env/v10"
)

// Config represents configuration model for the app environment.
type Config struct {
	// Address is web server host and port
	Address string `env:"ADDRESS"`
	// StoreInterval is interval in seconds to save current server metrics to disk
	StoreInterval int `env:"STORE_INTERVAL"`
	// FileStoragePath is full path to file where server metrics will be saved to/loaded from
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	// Restore is a flag to indicate whether to load previously saved metrics from file on server start or not
	Restore bool `env:"RESTORE"`
	// DatabaseDsn is a string to connect to the database
	DatabaseDsn string `env:"DATABASE_DSN"`
	// Key is a secret key used by the hash checker middleware
	Key string `env:"KEY"`
}

func ParseConfig() Config {
	var config Config

	flag.StringVar(&config.Address, "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")
	flag.IntVar(&config.StoreInterval, "i", 3000, "интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	flag.StringVar(&config.FileStoragePath, "f", "/tmp/metrics-db.json", "полное имя файла, куда сохраняются текущие значения")
	flag.BoolVar(&config.Restore, "r", true, "загружать или нет ранее сохранённые значения из указанного файла при старте сервера")
	flag.StringVar(&config.DatabaseDsn, "d", "", "Строка с адресом подключения к БД")
	flag.StringVar(&config.Key, "k", "", "Ключ")

	flag.Parse()

	err := env.Parse(&config)
	if err != nil {
		panic(err)
	}

	return config
}

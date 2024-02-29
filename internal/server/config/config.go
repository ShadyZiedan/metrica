package config

import (
	"flag"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDsn     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
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

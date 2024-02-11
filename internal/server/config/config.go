package config

import (
	"flag"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Address         string `env:"ADDRESS" envDefault:"localhost:8080"`
	StoreInterval   int    `env:"STORE_INTERVAL" envDefault:"3000"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"/tmp/metrics-db.json"`
	Restore         bool   `env:"RESTORE" envDefault:"true"`
}

func ParseConfig() Config {
	var config Config

	err := env.Parse(&config)
	if err != nil {
		panic(err)
	}

	flag.StringVar(&config.Address, "a", config.Address, "адрес эндпоинта HTTP-сервера")
	flag.IntVar(&config.StoreInterval, "i", config.StoreInterval, "интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	flag.StringVar(&config.FileStoragePath, "f", config.FileStoragePath, "полное имя файла, куда сохраняются текущие значения")
	flag.BoolVar(&config.Restore, "r", config.Restore, "загружать или нет ранее сохранённые значения из указанного файла при старте сервера")

	flag.Parse()
	return config
}

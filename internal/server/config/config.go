// Package config provides a way to parse and access configuration settings for the web server.
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v10"
	"github.com/shadyziedan/metrica/internal/server/logger"
	"go.uber.org/zap"
	"os"
	"time"
)

// Config represents configuration model for the app environment.
type Config struct {
	// Address is web server host and port
	Address string `env:"ADDRESS" json:"address"`
	// StoreInterval is interval in seconds to save current server metrics to disk
	StoreInterval Duration `env:"STORE_INTERVAL" json:"store_interval"`
	// FileStoragePath is full path to file where server metrics will be saved to/loaded from
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"store_file"`
	// Restore is a flag to indicate whether to load previously saved metrics from file on server start or not
	Restore bool `env:"RESTORE" json:"restore"`
	// DatabaseDsn is a string to connect to the database
	DatabaseDsn string `env:"DATABASE_DSN" json:"database_dsn"`
	// Key is a secret key used by the hash checker middleware
	Key string `env:"KEY" json:"-"`
	// CryptoKey is a path to the private key to decrypt message received from the agent
	CryptoKey string `env:"CRYPTO_KEY" json:"crypto_key"`
}

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

func ParseConfig() Config {
	var cnf Config

	var configPathJSON string
	flag.StringVar(&configPathJSON, "c", "", "Path to JSON config file")

	flag.StringVar(&cnf.Address, "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")
	flag.DurationVar(&cnf.StoreInterval.Duration, "i", 3*time.Second, "интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск")
	flag.StringVar(&cnf.FileStoragePath, "f", "/tmp/metrics-db.json", "полное имя файла, куда сохраняются текущие значения")
	flag.BoolVar(&cnf.Restore, "r", true, "загружать или нет ранее сохранённые значения из указанного файла при старте сервера")
	flag.StringVar(&cnf.DatabaseDsn, "d", "", "Строка с адресом подключения к БД")
	flag.StringVar(&cnf.Key, "k", "", "Ключ")
	flag.StringVar(&cnf.CryptoKey, "crypto-key", "", "путь до файла с приватным ключом")
	flag.Parse()

	if configPathJSON != "" {
		//var cnf2 Config
		if err := parseConfigFromJSONFile(configPathJSON, &cnf); err != nil {
			logger.Log.Fatal("failed to parse json config file", zap.Error(err))
		}
	}

	flag.Parse()
	if err := env.Parse(&cnf); err != nil {
		panic(err)
	}

	return cnf
}

func parseConfigFromJSONFile(jsonPath string, c *Config) error {
	f, err := os.Open(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(c)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	return nil
}

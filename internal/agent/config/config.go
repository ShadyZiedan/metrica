// Package config provides a way to parse and access configuration settings for the agent.
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"github.com/shadyziedan/metrica/internal/agent/logger"
	"go.uber.org/zap"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
)

// Config represents the configuration settings for the agent.
type Config struct {
	// Address is web server host address
	Address string `env:"ADDRESS" json:"address"`
	// ReportInterval is interval in seconds of reporting metrics
	ReportInterval Duration `env:"REPORT_INTERVAL" json:"report_interval"`
	// PollInterval is interval or frequency in seconds of gathering metrics from runtime package
	PollInterval Duration `env:"POLL_INTERVAL" json:"poll_interval"`
	// Key is a secret key for hashing data
	Key string `env:"KEY" json:"-"`
	// RateLimit is a rate limiter of metrics being sent
	RateLimit int `env:"RATE_LIMIT" json:"-"`
	// CryptoKey is a path to public key to encrypt data sent to the server
	CryptoKey string `env:"CRYPTO_KEY" json:"crypto_key"`
	// GrpcAddress is grpc server host address
	GrpcAddress string `env:"GRPC_ADDRESS" json:"grpc_address"`
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

// ParseConfig parses the command-line flags and environment variables to create a new Config instance.
func ParseConfig() Config {
	var cnf Config

	var configJSONPath string

	flag.StringVar(&configJSONPath, "c", "", "Path to JSON config file")
	flag.StringVar(&cnf.Address, "a", "localhost:8080", "адрес эндпоинта HTTP-сервера")
	flag.DurationVar(&cnf.ReportInterval.Duration, "r", 10*time.Second, "частота отправки метрик на сервер")
	flag.DurationVar(&cnf.PollInterval.Duration, "p", 2*time.Second, "частота опроса метрик из пакета runtime")
	flag.StringVar(&cnf.Key, "k", "", "Ключ")
	flag.IntVar(&cnf.RateLimit, "l", 1, "Rate limit")
	flag.StringVar(&cnf.CryptoKey, "crypto-key", "", "путь до файла с публичным ключом")
	flag.StringVar(&cnf.GrpcAddress, "grpc-address", "", "адрес grpc сервер")

	flag.Parse()

	if configJSONPath != "" {
		err := parseFromJSONFile(configJSONPath, &cnf)
		if err != nil {
			logger.Log.Fatal("failed to parse config from file", zap.Error(err))
		}
	}
	flag.Parse()

	err := env.Parse(&cnf)
	if err != nil {
		logger.Log.Fatal("failed to parse config from env", zap.Error(err))
	}

	return cnf
}

func parseFromJSONFile(configJSONPath string, config *Config) error {
	f, err := os.Open(configJSONPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(config)
}

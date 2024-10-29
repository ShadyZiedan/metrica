package main

import (
	"context"
	"github.com/shadyziedan/metrica/internal/security"
	"go.uber.org/zap"
	"os"
	"os/signal"

	"github.com/shadyziedan/metrica/internal/agent/agent"
	"github.com/shadyziedan/metrica/internal/agent/config"
	"github.com/shadyziedan/metrica/internal/agent/logger"
)

func main() {
	cnf := config.ParseConfig()
	err := logger.Initialize("info")
	if err != nil {
		panic(err)
	}
	var options []agent.Option
	if cnf.CryptoKey != "" {
		encryptor, err := security.NewDefaultEncryptorFromFile(cnf.CryptoKey)
		if err != nil {
			logger.Log.Error("encryption key is not loaded", zap.Error(err))
		} else {
			options = append(options, agent.WithEncryptor(encryptor))
		}
	}

	if cnf.Key != "" {
		hasher := security.NewDefaultHasher(cnf.Key)
		options = append(options, agent.WithHasher(hasher))
	}
	newAgent := agent.NewAgent(cnf, options...)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	newAgent.Run(ctx)
}

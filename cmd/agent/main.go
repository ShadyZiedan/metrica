package main

import (
	"context"
	"fmt"
	"github.com/shadyziedan/metrica/internal/security"
	"go.uber.org/zap"
	"os"
	"os/signal"

	"github.com/shadyziedan/metrica/internal/agent/agent"
	"github.com/shadyziedan/metrica/internal/agent/config"
	"github.com/shadyziedan/metrica/internal/agent/logger"
	"github.com/shadyziedan/metrica/internal/agent/services"
)

var (
	BuildVersion string
	BuildDate    string
	BuildCommit  string
)

func main() {
	showBuildInfo()
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
	newAgent := agent.NewAgent(cnf, services.NewMetricsCollector(), options...)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	newAgent.Run(ctx)
}

func showBuildInfo() {
	if BuildVersion != "" {
		fmt.Println("Build version: ", BuildVersion)
	} else {
		fmt.Println("Build version: N/A")
	}
	if BuildDate != "" {
		fmt.Println("Build date: ", BuildDate)
	} else {
		fmt.Println("Build date: N/A")
	}
	if BuildCommit != "" {
		fmt.Println("Build commit: ", BuildCommit)
	} else {
		fmt.Println("Build commit: N/A")
	}
}

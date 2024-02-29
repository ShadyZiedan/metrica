package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/shadyziedan/metrica/internal/agent/agent"
	"github.com/shadyziedan/metrica/internal/agent/config"
	"github.com/shadyziedan/metrica/internal/agent/logger"
)

func main() {
	conf := config.ParseConfig()
	err := logger.Initialize("info")
	if err != nil {
		panic(err)
	}
	newAgent := agent.NewAgent("http://"+conf.Address, conf.PollInterval, conf.ReportInterval, conf.Key, conf.RateLimit)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	newAgent.Run(ctx)
}

package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/shadyziedan/metrica/internal/agent/agent"
	"github.com/shadyziedan/metrica/internal/agent/config"
)

func main() {
	conf := config.ParseConfig()

	newAgent := agent.NewAgent("http://"+conf.Address, conf.PollInterval, conf.ReportInterval)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	newAgent.Run(ctx)
}

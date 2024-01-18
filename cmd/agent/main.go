package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/shadyziedan/metrica/cmd/agent/config"
	"github.com/shadyziedan/metrica/internal/agent/agent"
)

func main() {
	config := config.ParseConfig()

	agent := agent.NewAgent("http://"+config.Address, config.PollInterval, config.ReportInterval)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	agent.Run(ctx)
}

package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/shadyziedan/metrica/internal/agent/agent"
	"github.com/shadyziedan/metrica/internal/agent/config"
)

func main() {
	cnf := config.ParseConfig()

	newAgent := agent.NewAgent("http://"+cnf.Address, cnf.PollInterval, cnf.ReportInterval)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	newAgent.Run(ctx)
}

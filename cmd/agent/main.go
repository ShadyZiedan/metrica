package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/shadyziedan/metrica/internal/agent/agent"
	"github.com/shadyziedan/metrica/internal/agent/config"
	"github.com/shadyziedan/metrica/internal/agent/logger"
)

var (
	BuildVersion string
	BuildDate    string
	BuildCommit  string
)

func main() {
	showBuildInfo()
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

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/go-resty/resty/v2"
	"github.com/shadyziedan/metrica/internal/agent/services"
)

type Config struct {
	Address        string `env:"ADDRESS" envDefault:"localhost:8080"`
	ReportInterval int    `env:"REPORT_INTERVAL" envDefault:"10"`
	PollInterval   int    `env:"POLL_INTERVAL" envDefault:"2"`
}

func parseConfig() Config {
	var config Config

	err := env.Parse(&config)
	if err != nil {
		panic(err)
	}

	flag.StringVar(&config.Address, "a", config.Address, "адрес эндпоинта HTTP-сервера")
	flag.IntVar(&config.ReportInterval, "r", config.ReportInterval, "частота отправки метрик на сервер")
	flag.IntVar(&config.PollInterval, "p", config.PollInterval, "частота опроса метрик из пакета runtime")

	flag.Parse()
	return config
}

func main() {
	config := parseConfig()

	baseURL := "http://" + config.Address

	mc := services.NewMetricsCollector()
	go pollCollect(config.PollInterval, mc)
	go reportCollect(baseURL, config.ReportInterval, mc)
	c := make(chan os.Signal, 1)
	<-c
}

func reportCollect(baseURL string, reportInterval int, mc *services.MerticsCollector) {
	client := resty.New()
	client.BaseURL = baseURL
	for {
		metrics := mc.Collect()
		sendMetricsToServer(client, metrics)
		time.Sleep(time.Duration(reportInterval) * time.Second)
	}
}

func sendMetricsToServer(client *resty.Client, metrics *services.AgentMetrics) error {
	for metricName, val := range metrics.Gauge {
		_, err := client.R().Post(fmt.Sprintf("/update/gauge/%s/%v", metricName, val))
		if err != nil {
			return err
		}
	}
	for metricName, val := range metrics.Counter {
		_, err := client.R().Post(fmt.Sprintf("/update/counter/%s/%v", metricName, val))
		if err != nil {
			return err
		}
	}
	return nil
}

func pollCollect(pollInterval int, mc *services.MerticsCollector) {
	for {
		mc.Collect()
		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
}

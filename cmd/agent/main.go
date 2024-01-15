package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shadyziedan/metrica/internal/agent/services"
)

func main() {
	pollInterval := 2
	reportInterval := 10

	mc := services.NewMetricsCollector()
	baseURL := "http://localhost:8080"
	go pollCollect(pollInterval, mc)
	go reportCollect(baseURL, reportInterval, mc)
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

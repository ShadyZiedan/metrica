package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

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
	for {
		metrics := mc.Collect()
		sendMetricsToServer(baseURL, metrics)
		time.Sleep(time.Duration(reportInterval) * time.Second)
	}
}

func sendMetricsToServer(baseURL string, metrics *services.AgentMetrics) error {
	for metricName, val := range metrics.Gauge {
		err := sendGauge(baseURL, metricName, val)
		if err != nil {
			return err
		}
	}
	for metricName, val := range metrics.Counter {
		err := sendCounter(baseURL, metricName, val)
		if err != nil {
			return err
		}
	}
	return nil
}

func sendCounter(baseURL, metricName string, val int) error {
	url := fmt.Sprintf("%s/update/%s/%s/%v", baseURL, "counter", metricName, val)
	res, err := http.Post(url, "text/plain", nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.New("status not ok")
	}
	return nil
}

func sendGauge(baseURL, metricName string, val float64) error {
	url := fmt.Sprintf("%s/update/%s/%s/%v", baseURL, "gauge", metricName, val)
	res, err := http.Post(url, "text/plain", nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.New("status not ok")
	}
	return nil
}

func pollCollect(pollInterval int, mc *services.MerticsCollector) {
	for {
		mc.Collect()
		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
}

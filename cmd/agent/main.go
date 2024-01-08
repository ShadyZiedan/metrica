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
	baseUrl := "http://localhost:8080"
	go pollCollect(pollInterval, mc)
	go reportCollect(baseUrl, reportInterval, mc)
	c := make(chan os.Signal, 1)
	<- c
}

func reportCollect(baseUrl string, reportInterval int, mc *services.MerticsCollector) {
	for {
		metrics := mc.Collect()
		err := sendMetricsToServer(baseUrl, metrics)
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Duration(reportInterval) * time.Second)
	}
}


func sendMetricsToServer(baseUrl string, metrics *services.AgentMetrics) error {
	for metricName, val := range metrics.Gauge {
		url := fmt.Sprintf("%s/update/%s/%s/%v", baseUrl, "gauge", metricName, val)
		res, err := http.Post(url, "text/plain", nil)
		if err != nil {
			return err
		}
		if res.StatusCode != http.StatusOK {
			return errors.New("status not ok")
		}
	}
	for metricName, val := range metrics.Counter {
		url := fmt.Sprintf("%s/update/%s/%s/%v", baseUrl, "counter", metricName, val)
		res, err := http.Post(url, "text/plain", nil)
		if err != nil {
			return err
		}
		if res.StatusCode != http.StatusOK {
			return errors.New("status not ok")
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

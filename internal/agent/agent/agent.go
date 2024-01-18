package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shadyziedan/metrica/internal/agent/services"
)

type Agent struct {
	Client         *resty.Client
	PollInterval   int
	ReportInterval int
}

func NewAgent(baseURL string, pollInterval, reportInterval int) *Agent {
	client := resty.New()
	client.BaseURL = baseURL
	return &Agent{Client: client, PollInterval: pollInterval, ReportInterval: reportInterval}
}

func (a *Agent) Run(ctx context.Context) {

	mc := services.NewMetricsCollector()

	pollChan := time.NewTicker(time.Duration(a.PollInterval) * time.Second)
	reportChan := time.NewTicker(time.Duration(a.ReportInterval) * time.Second)

	for {
		select {
		case <-pollChan.C:
			mc.IncreasePollCount()
		case <-reportChan.C:
			metrics := mc.Collect()
			a.sendMetricsToServer(metrics)
		case <-ctx.Done():
			return
		}
	}
}

func (a *Agent) sendMetricsToServer(metrics *services.AgentMetrics) error {
	for metricName, val := range metrics.Gauge {
		_, err := a.Client.R().Post(fmt.Sprintf("/update/gauge/%s/%v", metricName, val))
		if err != nil {
			return err
		}
	}
	for metricName, val := range metrics.Counter {
		_, err := a.Client.R().Post(fmt.Sprintf("/update/counter/%s/%v", metricName, val))
		if err != nil {
			return err
		}
	}
	return nil
}

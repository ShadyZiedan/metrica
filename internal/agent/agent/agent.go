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

const ClientTimeout = 5

func NewAgent(baseURL string, pollInterval, reportInterval int) *Agent {
	client := resty.New()
	client.BaseURL = baseURL
	return &Agent{Client: client, PollInterval: pollInterval, ReportInterval: reportInterval}
}

func (a *Agent) Run(ctx context.Context) {

	mc := services.NewMetricsCollector()

	pollChan := time.NewTicker(time.Duration(a.PollInterval) * time.Second)
	defer pollChan.Stop()
	reportChan := time.NewTicker(time.Duration(a.ReportInterval) * time.Second)
	defer reportChan.Stop()

	for {
		select {
		case <-pollChan.C:
			mc.IncreasePollCount()
		case <-reportChan.C:
			metrics := mc.Collect()
			a.sendMetricsToServer(ctx, metrics)
		case <-ctx.Done():
			return
		}
	}
}

func (a *Agent) sendMetricsToServer(ctx context.Context, metrics *services.AgentMetrics) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, ClientTimeout*time.Second)
	defer cancel()
	for _, metric := range metrics.Gauge.GetAll() {
		_, err := a.Client.R().SetContext(timeoutCtx).Post(fmt.Sprintf("/update/gauge/%s/%v", metric.Name, metric.Value))
		if err != nil {
			return fmt.Errorf("update gauge '%s'->'%v': %w", metric.Name, metric.Value, err)
		}
	}
	for _, metric := range metrics.Counter.GetAll() {
		_, err := a.Client.R().SetContext(timeoutCtx).Post(fmt.Sprintf("/update/counter/%s/%v", metric.Name, metric.Value))
		if err != nil {
			return fmt.Errorf("update counter '%s'->'%v': %w", metric.Name, metric.Value, err)
		}
	}
	return nil
}

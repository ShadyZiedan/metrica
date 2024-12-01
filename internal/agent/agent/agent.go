// Package agent provides the functionality for a metric collection and reporting agent.
// It collects metrics from various sources, compresses and sends them to a server.
package agent

import (
	"context"
	"errors"
	"github.com/shadyziedan/metrica/internal/agent/config"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/shadyziedan/metrica/internal/agent/logger"
	"github.com/shadyziedan/metrica/internal/models"
	"github.com/shadyziedan/metrica/internal/retry"

	"github.com/shadyziedan/metrica/internal/agent/services"
)

// Agent represents a metric collection and reporting agent.
// It collects metrics from various sources, compresses and sends them to a server.
type Agent struct {
	PollInterval     time.Duration
	ReportInterval   time.Duration
	RateLimit        int
	metricsCollector metricsCollector
	metricsSender    metricsSender
}

type metricsSender interface {
	Send(ctx context.Context, metrics []*models.Metrics) error
}

type metricsCollector interface {
	Collect() *services.AgentMetrics
	IncreasePollCount()
}

// NewAgent creates a new instance of the Agent struct.
func NewAgent(cnf config.Config, mc metricsCollector, ms metricsSender) *Agent {
	a := &Agent{
		PollInterval:     cnf.PollInterval.Duration,
		ReportInterval:   cnf.ReportInterval.Duration,
		RateLimit:        cnf.RateLimit,
		metricsCollector: mc,
		metricsSender:    ms,
	}
	return a
}

// Run starts the metric collection and reporting process for the agent.
func (a *Agent) Run(ctx context.Context) {
	pollChan := time.NewTicker(a.PollInterval)
	defer pollChan.Stop()
	reportChan := time.NewTicker(a.ReportInterval)
	defer reportChan.Stop()

	metricsSendCh := make(chan *services.AgentMetrics, a.RateLimit)
	defer close(metricsSendCh)

	var wg sync.WaitGroup

	for i := 0; i < a.RateLimit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.sendMetricsWorker(ctx, metricsSendCh)
		}()
	}

	for {
		select {
		case <-pollChan.C:
			a.metricsCollector.IncreasePollCount()
		case <-reportChan.C:
			metrics := a.metricsCollector.Collect()
			metricsSendCh <- metrics
		case <-ctx.Done():
			wg.Wait()
			return
		}
	}
}

func (a *Agent) sendMetricsWorker(ctx context.Context, metricsCh <-chan *services.AgentMetrics) {
	for {
		select {
		case <-ctx.Done():
			return
		case metrics, ok := <-metricsCh:
			if !ok {
				return
			}
			err := retry.WithBackoff(ctx, 3, func(err error) bool {
				var e net.Error
				return errors.As(err, &e)
			}, func() error {
				return a.sendMetricsToServer(ctx, metrics)
			})
			if err != nil {
				logger.Log.Error("Error sending metric", zap.Error(err))
			}
		}
	}
}

func (a *Agent) sendMetricsToServer(ctx context.Context, metrics *services.AgentMetrics) error {
	var requestModels []*models.Metrics
	for _, metric := range metrics.Gauge.GetAll() {
		requestModels = append(requestModels, &models.Metrics{
			ID:    metric.Name,
			MType: "gauge",
			Value: &metric.Value,
		})
	}
	for _, metric := range metrics.Counter.GetAll() {
		delta := int64(metric.Value)
		requestModels = append(requestModels, &models.Metrics{
			ID:    metric.Name,
			MType: "counter",
			Delta: &delta,
		})
	}
	return a.metricsSender.Send(ctx, requestModels)
}

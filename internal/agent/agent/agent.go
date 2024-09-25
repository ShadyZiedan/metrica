package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/shadyziedan/metrica/internal/agent/logger"
	"github.com/shadyziedan/metrica/internal/models"
	"github.com/shadyziedan/metrica/internal/retry"
	"github.com/shadyziedan/metrica/internal/security"

	"github.com/go-resty/resty/v2"

	"github.com/shadyziedan/metrica/internal/agent/services"
)

// Agent represents a metric collection and reporting agent.
// It collects metrics from various sources, compresses and sends them to a server.
type Agent struct {
	Client         *resty.Client
	PollInterval   int
	ReportInterval int
	Key            string
	RateLimit      int
}

// NewAgent creates a new instance of the Agent struct.
func NewAgent(baseURL string, pollInterval, reportInterval int, key string, rateLimit int) *Agent {
	client := resty.New()
	client.BaseURL = baseURL
	return &Agent{Client: client, PollInterval: pollInterval, ReportInterval: reportInterval, Key: key, RateLimit: rateLimit}
}

// Run starts the metric collection and reporting process for the agent.
func (a *Agent) Run(ctx context.Context) {

	mc := services.NewMetricsCollector()

	pollChan := time.NewTicker(time.Duration(a.PollInterval) * time.Second)
	defer pollChan.Stop()
	reportChan := time.NewTicker(time.Duration(a.ReportInterval) * time.Second)
	defer reportChan.Stop()

	metricsSendCh := make(chan *services.AgentMetrics, a.RateLimit)
	defer close(metricsSendCh)

	var wg sync.WaitGroup

	for i := 0; i < a.RateLimit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sendMetricsWorker(ctx, a, metricsSendCh)
		}()
	}

	for {
		select {
		case <-pollChan.C:
			mc.IncreasePollCount()
		case <-reportChan.C:
			metrics := mc.Collect()
			metricsSendCh <- metrics
		case <-ctx.Done():
			wg.Wait()
			return
		}
	}
}

func sendMetricsWorker(ctx context.Context, a *Agent, metricsCh <-chan *services.AgentMetrics) {
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
	return a.sendMetrics(ctx, requestModels)
}

func (a *Agent) sendMetrics(ctx context.Context, metrics []*models.Metrics) error {
	body, err := convertMetricsToJSON(metrics)
	if err != nil {
		return fmt.Errorf("couldn't convert metrics to json string: %s", err)
	}
	bodyCompressed, err := compressBody(body)
	if err != nil {
		return err
	}
	req := a.Client.R().SetContext(ctx).SetBody(bodyCompressed).
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json")
	if a.Key != "" {
		hashHeader, err := security.Hash(bodyCompressed, a.Key)
		if err != nil {
			return err
		}
		req.SetHeader("HashSHA256", hashHeader)
	}
	_, err = req.Post("/updates/")
	return err
}

func convertMetricsToJSON(m []*models.Metrics) ([]byte, error) {
	jsonEncoded, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return jsonEncoded, nil
}
func compressBody(body []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	_, err = gzWriter.Write(body)
	if err != nil {
		return nil, err
	}
	if err = gzWriter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

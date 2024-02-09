package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shadyziedan/metrica/internal/models"

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
		model := &models.Metrics{
			ID:    metric.Name,
			MType: "gauge",
			Value: &metric.Value,
		}
		if err := a.sendMetric(timeoutCtx, model); err != nil {
			return fmt.Errorf("update gauge '%s'->'%v': %w", metric.Name, metric.Value, err)
		}
	}
	for _, metric := range metrics.Counter.GetAll() {
		delta := int64(metric.Value)
		model := &models.Metrics{
			ID:    metric.Name,
			MType: "counter",
			Delta: &delta,
		}
		if err := a.sendMetric(timeoutCtx, model); err != nil {
			return fmt.Errorf("update counter '%s'->'%v': %w", metric.Name, metric.Value, err)
		}
	}
	return nil
}

func (a Agent) sendMetric(ctx context.Context, model *models.Metrics) error {
	body, err := marshallAndCompressMetric(model)
	if err != nil {
		return fmt.Errorf("error marshalling and compressing model: %s", err)
	}
	_, err = a.Client.R().SetContext(ctx).
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetBody(body).Post("/update/")
	return err
}

func marshallAndCompressMetric(m *models.Metrics) ([]byte, error) {
	jsonEncoded, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	gzWriter, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	_, err = gzWriter.Write(jsonEncoded)
	if err != nil {
		return nil, err
	}
	if err = gzWriter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/go-errors/errors"
	"go.uber.org/zap"

	"github.com/shadyziedan/metrica/internal/agent/logger"
	"github.com/shadyziedan/metrica/internal/models"
	"github.com/shadyziedan/metrica/internal/retry"

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
	defer pollChan.Stop()
	reportChan := time.NewTicker(time.Duration(a.ReportInterval) * time.Second)
	defer reportChan.Stop()

	for {
		select {
		case <-pollChan.C:
			mc.IncreasePollCount()
		case <-reportChan.C:
			metrics := mc.Collect()
			err := retry.WithBackoff(ctx, 3, func(err error) bool {
				var e net.Error
				return errors.As(err, &e)
			}, func() error {
				return a.sendMetricsToServer(ctx, metrics)
			})
			if err != nil {
				logger.Log.Error("Error sending metric", zap.Error(err))
			}
		case <-ctx.Done():
			return
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
	body, err := marshallAndCompressMetrics(metrics)
	if err != nil {
		return fmt.Errorf("error marshalling and compressing model: %s", err)
	}
	_, err = a.Client.R().SetContext(ctx).
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetBody(body).Post("/updates/")
	return err
}

func marshallAndCompressMetrics(m []*models.Metrics) ([]byte, error) {
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

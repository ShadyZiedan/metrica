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
	"github.com/shadyziedan/metrica/internal/security"

	"github.com/go-resty/resty/v2"

	"github.com/shadyziedan/metrica/internal/agent/services"
)

type Agent struct {
	Client         *resty.Client
	PollInterval   int
	ReportInterval int
	Key            string
}

func NewAgent(baseURL string, pollInterval, reportInterval int, key string) *Agent {
	client := resty.New()
	client.BaseURL = baseURL
	return &Agent{Client: client, PollInterval: pollInterval, ReportInterval: reportInterval, Key: key}
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
	body, err := convertMetricsToJSON(metrics)
	if err != nil {
		return fmt.Errorf("error converting metrics to json string: %s", err)
	}
	bodyCompressed, err := compressBody(body)
	if err != nil {
		return err
	}
	req := a.Client.R().SetContext(ctx).SetBody(bodyCompressed).
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json")
	if a.Key != "" {
		hashHeader, err := security.Hash(body, a.Key)
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

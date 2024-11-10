// Package agent provides the functionality for a metric collection and reporting agent.
// It collects metrics from various sources, compresses and sends them to a server.
package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shadyziedan/metrica/internal/agent/config"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/go-resty/resty/v2"
	"github.com/shadyziedan/metrica/internal/agent/logger"
	"github.com/shadyziedan/metrica/internal/models"
	"github.com/shadyziedan/metrica/internal/retry"

	"github.com/shadyziedan/metrica/internal/agent/services"
)

// Agent represents a metric collection and reporting agent.
// It collects metrics from various sources, compresses and sends them to a server.
type Agent struct {
	Client           *resty.Client
	PollInterval     time.Duration
	ReportInterval   time.Duration
	RateLimit        int
	hasher           hasher
	encryptor        encryptor
	metricsCollector metricsCollector
	localIPAddress   string
}

type hasher interface {
	Hash(data []byte) (string, error)
}

type encryptor interface {
	Encrypt(data []byte) ([]byte, error)
	GetEncryptedKey() (string, error)
}

type Option = func(agent *Agent)

type metricsCollector interface {
	Collect() *services.AgentMetrics
	IncreasePollCount()
}

// NewAgent creates a new instance of the Agent struct.
func NewAgent(cnf config.Config, mc metricsCollector, options ...Option) *Agent {
	localIP, err := getLocalIPAddress()
	if err != nil {
		logger.Log.Fatal("Failed to get local IP address", zap.Error(err))
	}
	client := resty.New()
	client.BaseURL = cnf.Address
	a := &Agent{
		Client:           client,
		PollInterval:     cnf.PollInterval.Duration,
		ReportInterval:   cnf.ReportInterval.Duration,
		RateLimit:        cnf.RateLimit,
		metricsCollector: mc,
		localIPAddress:   localIP,
	}
	for _, option := range options {
		option(a)
	}
	return a
}

func WithHasher(hasher hasher) Option {
	return func(a *Agent) {
		a.hasher = hasher
	}
}

func WithEncryptor(encryptor encryptor) Option {
	return func(a *Agent) {
		a.encryptor = encryptor
	}
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
	return a.sendMetrics(ctx, requestModels)
}

func (a *Agent) sendMetrics(ctx context.Context, metrics []*models.Metrics) error {
	body, err := convertMetricsToJSON(metrics)
	if err != nil {
		return fmt.Errorf("couldn't convert metrics to json string: %s", err)
	}

	req := a.Client.R().SetContext(ctx).
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetHeader("X-REAL-IP", a.localIPAddress)

	// Encrypt the json body
	if a.encryptor != nil {
		encryptedKey, encryptionError := a.encryptor.GetEncryptedKey()
		if encryptionError != nil {
			return fmt.Errorf("error encrypting metrics data: %s", encryptionError)
		}
		req.SetHeader(`X-Encrypted-Key`, encryptedKey)

		bodyEncrypted, encryptionError := a.encryptor.Encrypt(body)
		if encryptionError != nil {
			return fmt.Errorf("error encrypting metrics data: %s", encryptionError)
		}

		body = []byte(base64.StdEncoding.EncodeToString(bodyEncrypted))
	}

	// Compress the data
	bodyCompressed, err := compressBody(body)
	if err != nil {
		return err
	}

	if a.hasher != nil {
		hashHeader, hashErr := a.hasher.Hash(bodyCompressed)
		if hashErr != nil {
			return hashErr
		}
		req.SetHeader("HashSHA256", hashHeader)
	}

	res, err := req.SetBody(bodyCompressed).Post("/updates/")
	if err != nil {
		return fmt.Errorf("couldn't send metrics: %w", err)
	}
	if res.IsError() {
		return fmt.Errorf("request failed with status %d: %s", res.StatusCode(), res.String())
	}
	return nil
}

func getLocalIPAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// Проверяем, что это не loopback-адрес и что это IPv4
			if ip != nil && !ip.IsLoopback() && ip.To4() != nil {
				return ip.String(), nil
			}
		}
	}
	return "", fmt.Errorf("local IP address not found")
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

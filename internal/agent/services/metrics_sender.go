package services

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/shadyziedan/metrica/internal/models"
	"github.com/shadyziedan/metrica/internal/server/logger"
	pb "github.com/shadyziedan/metrica/proto"
	"go.uber.org/zap"
	"net"
)

type MetricsSender struct {
	httpClient     *resty.Client
	grpcClient     pb.MetricsServiceClient
	hasher         hasher
	encryptor      encryptor
	localIPAddress string
}

type MetricsSenderOption = func(agent *MetricsSender)

func NewMetricsSender(httpClient *resty.Client, grpcClient pb.MetricsServiceClient, options ...MetricsSenderOption) *MetricsSender {
	sender := &MetricsSender{httpClient: httpClient, grpcClient: grpcClient}
	for _, option := range options {
		option(sender)
	}
	if httpClient != nil {
		localIPAddress, err := getLocalIPAddress()
		if err != nil {
			logger.Log.Fatal("failed to get local ip address", zap.Error(err))
		}
		sender.localIPAddress = localIPAddress
	}
	return sender
}

type hasher interface {
	Hash(data []byte) (string, error)
}

type encryptor interface {
	Encrypt(data []byte) ([]byte, error)
	GetEncryptedKey() (string, error)
}

func WithHasher(hasher hasher) MetricsSenderOption {
	return func(a *MetricsSender) {
		a.hasher = hasher
	}
}

func WithEncryptor(encryptor encryptor) MetricsSenderOption {
	return func(a *MetricsSender) {
		a.encryptor = encryptor
	}
}

func (s *MetricsSender) Send(ctx context.Context, metrics []*models.Metrics) error {
	if s.grpcClient != nil {
		return s.sendGRPC(ctx, metrics)
	} else if s.httpClient != nil {
		return s.SendJSON(ctx, metrics)
	}
	return fmt.Errorf("no grpc client or http client")
}

func (s *MetricsSender) sendGRPC(ctx context.Context, metrics []*models.Metrics) error {
	requestModels := make([]*pb.Metric, len(metrics))
	for i, m := range metrics {
		var mType pb.Metric_MType
		switch m.MType {
		case "counter":
			mType = pb.Metric_COUNTER
		case "gauge":
			mType = pb.Metric_GAUGE
		}
		requestModels[i] = &pb.Metric{
			Id:    m.ID,
			MType: mType,
			Delta: m.Delta,
			Value: m.Value,
		}
	}
	_, err := s.grpcClient.UpdateBatchMetrics(ctx, &pb.UpdateBatchRequest{Metrics: requestModels})
	if err != nil {
		return err
	}
	return nil
}

func (s *MetricsSender) SendJSON(ctx context.Context, metrics []*models.Metrics) error {
	body, err := convertMetricsToJSON(metrics)
	if err != nil {
		return fmt.Errorf("couldn't convert metrics to json string: %s", err)
	}

	req := s.httpClient.R().SetContext(ctx).
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetHeader("X-REAL-IP", s.localIPAddress)

	// Encrypt the json body
	if s.encryptor != nil {
		encryptedKey, encryptionError := s.encryptor.GetEncryptedKey()
		if encryptionError != nil {
			return fmt.Errorf("error encrypting metrics data: %s", encryptionError)
		}
		req.SetHeader(`X-Encrypted-Key`, encryptedKey)

		bodyEncrypted, encryptionError := s.encryptor.Encrypt(body)
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

	if s.hasher != nil {
		hashHeader, hashErr := s.hasher.Hash(bodyCompressed)
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

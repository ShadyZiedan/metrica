package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"go.uber.org/zap"

	"github.com/shadyziedan/metrica/internal/models"
	"github.com/shadyziedan/metrica/internal/retry"
	"github.com/shadyziedan/metrica/internal/server/logger"
)

type FileStorage struct {
	producer *producer
	consumer *consumer
	fileName string
	mode     Mode

	mutex sync.RWMutex
}

func NewFileStorage(fileName string, mode Mode) *FileStorage {
	return &FileStorage{fileName: fileName, mode: mode}
}

type producer struct {
	file    *os.File
	encoder *json.Encoder
}

type consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func newConsumer(fileName string) (*consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &consumer{file: file, decoder: json.NewDecoder(file)}, nil
}

type Mode string

const (
	Sync   Mode = "sync"
	Normal Mode = "normal"
)

func newProducer(fileName string, mode Mode) (*producer, error) {
	var flag int
	switch mode {
	case Sync:
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND | os.O_SYNC
	case Normal:
		flag = os.O_WRONLY | os.O_CREATE | os.O_SYNC
	}
	var file *os.File
	err := retry.WithBackoff(context.Background(), 3, func(err error) bool {
		return err != nil
	}, func() error {
		var openErr error
		file, openErr = os.OpenFile(fileName, flag, 0666)
		return openErr
	})
	if err != nil {
		return nil, err
	}
	return &producer{file: file, encoder: json.NewEncoder(file)}, nil
}

func (fs *FileStorage) ReadMetrics() ([]*models.Metrics, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	if err := fs.OpenConsumer(); err != nil {
		return nil, err
	}
	defer func() {
		if err := fs.CloseConsumer(); err != nil {
			logger.Log.Info("close consumer error", zap.Error(err))
		}
	}()
	var res []*models.Metrics
	for {
		model := &models.Metrics{}
		err := fs.consumer.decoder.Decode(&model)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		res = append(res, model)
	}
	return res, nil
}

func (fs *FileStorage) SaveMetric(metric *models.Metrics) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	err := fs.OpenProducer()
	if err != nil {
		return err
	}
	defer func() {
		if err := fs.CloseProducer(); err != nil {
			logger.Log.Error("close producer failed", zap.Error(err))
		}
	}()
	return fs.producer.encoder.Encode(metric)
}

func (fs *FileStorage) SaveMetrics(metrics []*models.Metrics) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	if err := fs.OpenProducer(); err != nil {
		return err
	}
	defer func() {
		if err := fs.CloseProducer(); err != nil {
			panic(err)
		}
	}()
	for _, metric := range metrics {
		if err := fs.producer.encoder.Encode(metric); err != nil {
			return err
		}
	}
	return nil
}

func (c *consumer) Close() error {
	return c.file.Close()
}

func (p *producer) Close() error {
	return p.file.Close()
}

func (fs *FileStorage) CloseProducer() error {
	if fs.producer == nil {
		return nil
	}

	defer func() { fs.producer = nil }()

	if err := fs.producer.file.Sync(); err != nil {
		return fmt.Errorf("failed to flush producer file: %w", err)
	}

	if err := fs.producer.file.Close(); err != nil {
		return fmt.Errorf("failed to close producer file: %w", err)
	}
	return nil
}
func (fs *FileStorage) CloseConsumer() error {
	if fs.consumer == nil {
		return nil
	}

	defer func() { fs.consumer = nil }()

	if err := fs.consumer.file.Sync(); err != nil {
		return fmt.Errorf("failed to flush consumer file: %w", err)
	}

	if err := fs.consumer.file.Close(); err != nil {
		return fmt.Errorf("failed to close consumer file: %w", err)
	}

	return nil
}

func (fs *FileStorage) Close() error {
	if err := fs.CloseProducer(); err != nil {
		return err
	}
	return fs.CloseConsumer()
}

func (fs *FileStorage) OpenProducer() error {
	p, err := newProducer(fs.fileName, fs.mode)
	if err != nil {
		return err
	}
	fs.producer = p
	return nil
}

func (fs *FileStorage) OpenConsumer() error {
	c, err := newConsumer(fs.fileName)
	if err != nil {
		return err
	}
	fs.consumer = c
	return nil
}

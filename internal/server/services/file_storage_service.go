package services

import (
	"context"
	"github.com/shadyziedan/metrica/internal/models"
	"github.com/shadyziedan/metrica/internal/server/storage"
	"io"
	"time"
)

type metricsRepository interface {
	FindOrCreate(name string) (*models.Metric, error)
	FindAll() ([]*models.Metric, error)
}

type fileStorage interface {
	OpenProducer() error
	OpenConsumer() error
	io.Closer
	ReadMetrics() ([]*models.Metrics, error)
	SaveMetric(metric *models.Metrics) error
}

type FileStorageService struct {
	conf              FileStorageServiceConfig
	fileStorage       fileStorage
	metricsRepository metricsRepository
}

func NewFileStorageService(metricsRepository metricsRepository, conf FileStorageServiceConfig) *FileStorageService {
	fs := storage.NewFileStorage(conf.FileStoragePath)
	return &FileStorageService{conf: conf, fileStorage: fs, metricsRepository: metricsRepository}
}

type FileStorageServiceConfig struct {
	FileStoragePath string
	StoreInterval   int
	Restore         bool
}

func (s *FileStorageService) Run(ctx context.Context) {
	if s.conf.Restore {
		err := s.restoreRepository()
		if err != nil {
			panic(err)
		}
	}
	updateStorageTicker := time.NewTicker(time.Duration(s.conf.StoreInterval) * time.Second)
	defer updateStorageTicker.Stop()

	for {
		select {
		case <-updateStorageTicker.C:
			err := s.updateStorage()
			if err != nil {
				panic(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *FileStorageService) updateStorage() error {
	err := s.fileStorage.OpenProducer()
	if err != nil {
		return err
	}
	defer s.fileStorage.Close()
	metrics, err := s.metricsRepository.FindAll()
	if err != nil {
		return err
	}
	for _, metric := range metrics {
		model := &models.Metrics{
			ID:    metric.Name,
			MType: metric.MType,
			Delta: &metric.Counter,
			Value: &metric.Gauge,
		}
		err = s.fileStorage.SaveMetric(model)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *FileStorageService) restoreRepository() error {
	err := s.fileStorage.OpenConsumer()
	if err != nil {
		return err
	}
	defer s.fileStorage.Close()
	metrics, err := s.fileStorage.ReadMetrics()
	if err != nil {
		return err
	}
	for _, metric := range metrics {
		model, err := s.metricsRepository.FindOrCreate(metric.ID)
		if err != nil {
			return err
		}
		model.MType = metric.MType
		model.Gauge = *metric.Value
		model.Counter = *metric.Delta
	}
	return nil
}

package services

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/shadyziedan/metrica/internal/models"
	"github.com/shadyziedan/metrica/internal/server/logger"
	"github.com/shadyziedan/metrica/internal/server/storage"
)

type metricsRepository interface {
	FindOrCreate(ctx context.Context, name string, mType string) (*models.Metric, error)
	FindAll(ctx context.Context) ([]*models.Metric, error)
	Attach(observer storage.MetricsObserver)
	Detach(observer storage.MetricsObserver)
}

type fileStorage interface {
	ReadMetrics() ([]*models.Metrics, error)
	SaveMetric(*models.Metrics) error
	SaveMetrics([]*models.Metrics) error
}

type FileStorageService struct {
	conf              FileStorageServiceConfig
	fileStorage       fileStorage
	metricsRepository metricsRepository
}

func NewFileStorageService(metricsRepository metricsRepository, conf FileStorageServiceConfig) *FileStorageService {
	var mode storage.Mode
	if conf.StoreInterval == 0 {
		mode = storage.Sync
	} else {
		mode = storage.Normal
	}
	fs := storage.NewFileStorage(conf.FileStoragePath, mode)
	return &FileStorageService{conf: conf, fileStorage: fs, metricsRepository: metricsRepository}
}

type FileStorageServiceConfig struct {
	FileStoragePath string
	StoreInterval   int
	Restore         bool
}

func (s *FileStorageService) Run(ctx context.Context) {
	if s.conf.Restore {
		err := s.restoreRepository(ctx)
		if err != nil {
			panic(err)
		}
	}

	if s.conf.StoreInterval == 0 { //Sync mode
		s.Observe()
		<-ctx.Done()
		s.StopObserving()
		return
	}

	updateStorageTicker := time.NewTicker(time.Duration(s.conf.StoreInterval) * time.Second)
	defer updateStorageTicker.Stop()

	for {
		select {
		case <-updateStorageTicker.C:
			err := s.updateStorage(ctx)
			if err != nil {
				logger.Log.Error("Failed to update metrics storage", zap.Error(err))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *FileStorageService) Notify(metric *models.Metric) error {
	model := &models.Metrics{
		ID:    metric.Name,
		MType: metric.MType,
		Delta: metric.Counter,
		Value: metric.Gauge,
	}
	err := s.fileStorage.SaveMetric(model)
	if err != nil {
		return err
	}
	return nil
}

func (s *FileStorageService) updateStorage(ctx context.Context) error {
	metrics, err := s.metricsRepository.FindAll(ctx)
	if err != nil {
		return err
	}
	jsonModels := make([]*models.Metrics, 0, len(metrics))
	for _, metric := range metrics {
		model := &models.Metrics{
			ID:    metric.Name,
			MType: metric.MType,
			Delta: metric.Counter,
			Value: metric.Gauge,
		}
		jsonModels = append(jsonModels, model)
	}
	return s.fileStorage.SaveMetrics(jsonModels)
}

func (s *FileStorageService) restoreRepository(ctx context.Context) error {
	metrics, err := s.fileStorage.ReadMetrics()
	if err != nil {
		return err
	}
	for _, metric := range metrics {
		model, err := s.metricsRepository.FindOrCreate(ctx, metric.ID, metric.MType)
		if err != nil {
			return err
		}
		model.MType = metric.MType
		model.Gauge = metric.Value
		model.Counter = metric.Delta
	}
	return nil
}

func (s *FileStorageService) Observe() {
	s.metricsRepository.Attach(s)
}

func (s *FileStorageService) StopObserving() {
	s.metricsRepository.Detach(s)
}

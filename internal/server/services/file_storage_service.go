package services

import (
	"context"
	"time"

	"github.com/shadyziedan/metrica/internal/models"
	"github.com/shadyziedan/metrica/internal/server/storage"
)

type metricsRepository interface {
	FindOrCreate(name string) (*models.Metric, error)
	FindAll() ([]*models.Metric, error)
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
	observer          chan *models.Metric
}

func NewFileStorageService(metricsRepository metricsRepository, conf FileStorageServiceConfig) *FileStorageService {
	var mode storage.Mode
	if conf.StoreInterval == 0 {
		mode = storage.Async
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
		err := s.restoreRepository()
		if err != nil {
			panic(err)
		}
	}

	if s.conf.StoreInterval == 0 { //Async mode
		observer := s.Observe()
		defer s.Stop()
		for {
			select {
			case metric := <-observer:
				err := s.saveMetric(metric)
				if err != nil {
					panic(err)
				}
			case <-ctx.Done():
				return
			}
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

func (s *FileStorageService) saveMetric(metric *models.Metric) error {
	model := &models.Metrics{
		ID:    metric.Name,
		MType: metric.MType,
		Delta: &metric.Counter,
		Value: &metric.Gauge,
	}
	err := s.fileStorage.SaveMetric(model)
	if err != nil {
		return err
	}
	return nil
}

func (s *FileStorageService) updateStorage() error {
	metrics, err := s.metricsRepository.FindAll()
	if err != nil {
		return err
	}
	jsonModels := make([]*models.Metrics, 0, len(metrics))
	for _, metric := range metrics {
		model := &models.Metrics{
			ID:    metric.Name,
			MType: metric.MType,
			Delta: &metric.Counter,
			Value: &metric.Gauge,
		}
		jsonModels = append(jsonModels, model)
	}
	return s.fileStorage.SaveMetrics(jsonModels)
}

func (s *FileStorageService) restoreRepository() error {
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

func (s *FileStorageService) Observe() <-chan *models.Metric {
	c := make(chan *models.Metric)
	s.metricsRepository.Attach(c)
	s.observer = c
	return c
}

func (s *FileStorageService) Stop() {
	close(s.observer)
}

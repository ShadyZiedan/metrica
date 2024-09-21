package storage

import (
	"context"
	"slices"
	"sync"

	"github.com/go-errors/errors"
	"golang.org/x/exp/maps"

	"github.com/shadyziedan/metrica/internal/models"
)

type MemStorage struct {
	storage          map[string]*models.Metric
	m                sync.RWMutex
	metricsObservers []MetricsObserver
}

type MetricsObserver interface {
	Notify(*models.Metric) error
}

func (s *MemStorage) Attach(observer MetricsObserver) {
	s.metricsObservers = append(s.metricsObservers, observer)
}

func (s *MemStorage) Detach(observer MetricsObserver) {
	s.metricsObservers = slices.DeleteFunc(s.metricsObservers, func(o MetricsObserver) bool {
		return o == observer
	})
}

var (
	ErrMetricNotCreated    = errors.New("couldn't create a metric")
	ErrMetricAlreadyExists = errors.New("metric has been already created")
	ErrMetricNotFound      = errors.New("metric not found")
)

// FindAll implements MetricsRepository.
func (s *MemStorage) FindAll(ctx context.Context) ([]*models.Metric, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	return maps.Values(s.storage), nil
}

// FindOrCreate implements MetricsRepository.
func (s *MemStorage) FindOrCreate(ctx context.Context, name string, mType string) (*models.Metric, error) {
	if metric, err := s.Find(ctx, name); err == nil {
		return metric, nil
	}
	if err := s.Create(ctx, name, mType); err != nil {
		return nil, ErrMetricNotCreated
	}
	return s.Find(ctx, name)
}

// Create implements MetricsRepository.
func (s *MemStorage) Create(ctx context.Context, name string, mType string) error {
	s.m.Lock()
	defer s.m.Unlock()
	if _, err := s.Find(ctx, name); err == nil {
		return ErrMetricAlreadyExists
	}
	s.storage[name] = &models.Metric{Name: name, MType: mType}
	return nil
}

// Find implements MetricsRepository.
func (s *MemStorage) Find(ctx context.Context, name string) (*models.Metric, error) {
	if v, ok := s.storage[name]; ok {
		return v, nil
	}
	return nil, ErrMetricNotFound
}

func (s *MemStorage) UpdateCounter(ctx context.Context, name string, delta int64) error {
	model, err := s.Find(ctx, name)
	if err != nil {
		return err
	}
	model.MType = "counter"
	model.UpdateCounter(delta)
	return s.notify(ctx, model)
}

func (s *MemStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	model, err := s.Find(ctx, name)
	if err != nil {
		return err
	}
	model.MType = "gauge"
	model.UpdateGauge(value)
	return s.notify(ctx, model)
}

func (s *MemStorage) FindAllByName(ctx context.Context, names []string) ([]*models.Metric, error) {
	metrics, err := s.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, metric := range metrics {
		for _, name := range names {
			if metric.Name == name {
				metrics = append(metrics, metric)
			}
		}
	}
	return metrics, nil
}

func (s *MemStorage) notify(ctx context.Context, model *models.Metric) error {
	for _, metricsObserver := range s.metricsObservers {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := metricsObserver.Notify(model); err != nil {
				return err
			}
		}
	}
	return nil
}

func NewMemStorage() *MemStorage {
	return &MemStorage{storage: make(map[string]*models.Metric)}
}

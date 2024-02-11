package storage

import (
	"sync"

	"github.com/go-errors/errors"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

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
	slices.DeleteFunc(s.metricsObservers, func(o MetricsObserver) bool {
		return o == observer
	})
}

var (
	ErrMetricNotCreated    = errors.New("couldn't create a metric")
	ErrMetricAlreadyExists = errors.New("metric has been already created")
	ErrMetricNotFound      = errors.New("metric not found")
)

// FindAll implements MetricsRepository.
func (s *MemStorage) FindAll() ([]*models.Metric, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	return maps.Values(s.storage), nil
}

// FindOrCreate implements MetricsRepository.
func (s *MemStorage) FindOrCreate(name string) (*models.Metric, error) {
	if metric, err := s.Find(name); err == nil {
		return metric, nil
	}
	if err := s.Create(name); err != nil {
		return nil, ErrMetricNotCreated
	}
	return s.Find(name)
}

// Create implements MetricsRepository.
func (s *MemStorage) Create(name string) error {
	s.m.Lock()
	defer s.m.Unlock()
	if _, err := s.Find(name); err == nil {
		return ErrMetricAlreadyExists
	}
	s.storage[name] = &models.Metric{Name: name}
	return nil
}

// Find implements MetricsRepository.
func (s *MemStorage) Find(name string) (*models.Metric, error) {
	if v, ok := s.storage[name]; ok {
		return v, nil
	}
	return nil, ErrMetricNotFound
}

func (s *MemStorage) UpdateCounter(name string, delta int64) error {
	model, err := s.Find(name)
	if err != nil {
		return err
	}
	model.MType = "counter"
	model.UpdateCounter(delta)
	return s.notify(model)
}

func (s *MemStorage) UpdateGauge(name string, value float64) error {
	model, err := s.Find(name)
	if err != nil {
		return err
	}
	model.MType = "gauge"
	model.UpdateGauge(value)
	return s.notify(model)
}

func (s *MemStorage) notify(model *models.Metric) error {
	for _, metricsObserver := range s.metricsObservers {
		err := metricsObserver.Notify(model)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewMemStorage() *MemStorage {
	return &MemStorage{storage: make(map[string]*models.Metric)}
}

package storage

import (
	"sync"

	"github.com/go-errors/errors"
	"github.com/shadyziedan/metrica/internal/models"
	"golang.org/x/exp/maps"
)

type MemStorage struct {
	storage map[string]*models.Metric
	m       sync.RWMutex
}

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

// Create implements MetricsStorage.
func (s *MemStorage) Create(name string) error {
	s.m.Lock()
	defer s.m.Unlock()
	if _, err := s.Find(name); err == nil {
		return errors.New("Metric has been already created")
	}
	s.storage[name] = &models.Metric{Name: name}
	return nil
}

// Find implements MetricsStorage.
func (s *MemStorage) Find(name string) (*models.Metric, error) {
	if v, ok := s.storage[name]; ok {
		return v, nil
	}
	return nil, errors.New("Metric not found")
}

var _ MetricsRepository = (*MemStorage)(nil)

func NewMemStorage() *MemStorage {
	return &MemStorage{storage: make(map[string]*models.Metric)}
}

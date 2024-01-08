package storage

import (
	"github.com/go-errors/errors"
	"github.com/shadyziedan/metrica/internal/models"
)

type MemStorage struct {
	storage map[string]*models.Metric
}

// FindOrCreate implements MetricsRepository.
func (s *MemStorage) FindOrCreate(name string) (*models.Metric, error) {
	if metric, err := s.Find(name); err == nil {
		return metric, nil
	}
	if err := s.Create(name); err != nil {
		return nil, errors.New("Couldn't create a metric")
	}
	return s.Find(name)
}

// Create implements MetricsStorage.
func (s *MemStorage) Create(name string) error {
	if _, err := s.Find(name); err == nil {
		return errors.New("Metric has been already created")
	}
	s.storage[name] = &models.Metric{}
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

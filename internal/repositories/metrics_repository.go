package repositories

import (
	"errors"

	"github.com/shadyziedan/metrica/internal/models"
)

type MetricsRepository interface {
	Find(name string) (*models.Metric, error)
	Create(name string) error
	FindOrCreate(name string) (*models.Metric, error)
	FindAll() ([]*models.Metric, error)
}

var ErrMetricNotCreated = errors.New("couldn't create a metric")

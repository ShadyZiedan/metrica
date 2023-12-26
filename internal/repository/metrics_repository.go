package repository

import "github.com/shadyziedan/metrica/internal/models"

type MetricsRepository interface {
	Find(name string) (*models.Metric, error)
	Create(name string) error
	FindOrCreate(name string) (*models.Metric, error)
}

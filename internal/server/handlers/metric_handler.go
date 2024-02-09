package handlers

import "github.com/shadyziedan/metrica/internal/models"

type MetricHandler struct {
	repository metricsRepository
}

type metricsRepository interface {
	Find(name string) (*models.Metric, error)
	Create(name string) error
	FindOrCreate(name string) (*models.Metric, error)
	FindAll() ([]*models.Metric, error)
	UpdateCounter(name string, delta int64) error
	UpdateGauge(name string, value float64) error
}

func NewMetricHandler(repository metricsRepository) *MetricHandler {
	return &MetricHandler{repository: repository}
}

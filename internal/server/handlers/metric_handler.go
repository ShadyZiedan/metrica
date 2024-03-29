package handlers

import (
	"context"

	"github.com/shadyziedan/metrica/internal/models"
)

type MetricHandler struct {
	repository metricsRepository
	conn       dbConnection
}

type dbConnection interface {
	Ping(ctx context.Context) error
}

type metricsRepository interface {
	Find(ctx context.Context, name string) (*models.Metric, error)
	Create(ctx context.Context, name string, mType string) error
	FindOrCreate(ctx context.Context, name string, mType string) (*models.Metric, error)
	FindAll(ctx context.Context) ([]*models.Metric, error)
	UpdateCounter(ctx context.Context, name string, delta int64) error
	UpdateGauge(ctx context.Context, name string, value float64) error
	FindAllByName(ctx context.Context, names []string) ([]*models.Metric, error)
}

func NewMetricHandler(conn dbConnection, repository metricsRepository) *MetricHandler {
	return &MetricHandler{repository: repository, conn: conn}
}

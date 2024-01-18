package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/shadyziedan/metrica/internal/repositories"
)

func NewRouter(repo repositories.MetricsRepository) chi.Router {
	r := chi.NewRouter()
	r.Post(`/update/{metricType}/{metricName}/{metricValue}`, UpdateMetricHandler(repo))
	r.Get(`/value/{metricType}/{metricName}`, MetricHandler(repo))
	r.Get(`/`, MetricsHandler(repo))
	return r
}

package handlers

import (
	"github.com/go-chi/chi/v5"
)

func NewRouter(repo metricsRepository) chi.Router {
	r := chi.NewRouter()
	metricsHandler := NewMetricHandler(repo)
	r.Post(`/update/{metricType}/{metricName}/{metricValue}`, metricsHandler.UpdateMetricHandler)
	r.Get(`/value/{metricType}/{metricName}`, metricsHandler.GetMetric)
	r.Get(`/`, metricsHandler.GetAll)
	return r
}

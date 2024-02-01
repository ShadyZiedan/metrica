package handlers

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

type middleware = func(http.Handler) http.Handler

func NewRouter(repo metricsRepository, middlewares ...middleware) chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares...)
	metricsHandler := NewMetricHandler(repo)
	r.Post(`/update/{metricType}/{metricName}/{metricValue}`, metricsHandler.UpdateMetricHandler)
	r.Get(`/value/{metricType}/{metricName}`, metricsHandler.GetMetric)
	r.Get(`/`, metricsHandler.GetAll)

	//json api
	r.Post(`/update/`, metricsHandler.UpdateJson)
	r.Post(`/value/`, metricsHandler.GetMetricJson)
	return r
}

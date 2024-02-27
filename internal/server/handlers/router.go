package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type middleware = func(http.Handler) http.Handler

func NewRouter(conn dbConnection, repo metricsRepository, middlewares ...middleware) chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares...)
	metricsHandler := NewMetricHandler(conn, repo)
	r.Post(`/update/{metricType}/{metricName}/{metricValue}`, metricsHandler.UpdateMetricHandler)
	r.Get(`/value/{metricType}/{metricName}`, metricsHandler.GetMetric)
	r.Get(`/`, metricsHandler.GetAll)
	r.Get(`/ping`, metricsHandler.Ping)

	//json api
	r.Post(`/update/`, metricsHandler.UpdateJSON)
	r.Post(`/value/`, metricsHandler.GetMetricJSON)
	r.Post(`/updates/`, metricsHandler.UpdateBatch)
	return r
}

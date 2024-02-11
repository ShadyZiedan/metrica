package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type middleware = func(http.Handler) http.Handler

func NewRouter(conn *pgx.Conn, repo metricsRepository, middlewares ...middleware) chi.Router {
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
	return r
}

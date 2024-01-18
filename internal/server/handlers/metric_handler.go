package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shadyziedan/metrica/internal/repositories"
)

func MetricHandler(repo repositories.MetricsRepository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metric, err := repo.Find(metricName)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}
		switch metricType {
		case "counter":
			io.WriteString(rw, fmt.Sprintf("%v", metric.GetCounter()))
			return
		case "gauge":
			io.WriteString(rw, fmt.Sprintf("%v", metric.GetGauge()))
			return
		default:
			http.Error(rw, "unknown metric type: "+metricType, http.StatusNotFound)
			return
		}
	}
}

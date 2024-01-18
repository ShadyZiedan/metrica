package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/shadyziedan/metrica/internal/repositories"
)

func UpdateMetricHandler(storage repositories.MetricsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		metric, err := storage.FindOrCreate(metricName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		switch metricType {
		case "counter":
			num, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			metric.UpdateCounter(num)
			return
		case "gauge":
			num, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			metric.UpdateGauge(num)
			return
		default:
			http.Error(w, "unknown metric type", http.StatusBadRequest)
		}

	}
}

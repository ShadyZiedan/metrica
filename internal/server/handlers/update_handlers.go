package handlers

import (
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/shadyziedan/metrica/internal/server/storage"
)

func UnknownMetricHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "unknown metric type", http.StatusBadRequest)
}

func UpdateMetricHandler(storage storage.MetricsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")
		if len(parts) != 3 {
			http.NotFound(w, r)
			return
		}
		metricType := parts[0]
		metricName := parts[1]
		metricValue := parts[2]
		if !slices.Contains([]string{"counter", "gauge"}, metricType) {
			http.Error(w, "unknown metric type", http.StatusInternalServerError)
			return
		}
		metric, err := storage.FindOrCreate(metricName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		switch metricType {
		case "counter":
			num , err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			metric.UpdateCounter(num)
			return
		case "gauge":
			num , err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			metric.UpdateGauge(num)
			return
		default:
			http.Error(w, "unknown metric type", http.StatusInternalServerError)
		}

	}
}
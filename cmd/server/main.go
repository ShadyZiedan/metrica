package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/shadyziedan/metrica/internal/repository"
)

func main() {
	storage := repository.NewMemStorage()
	mux := http.NewServeMux()
	mux.Handle(`/update/counter/`, UpdateMetricHandler("counter", storage))
	mux.Handle(`/update/gauge/`, UpdateMetricHandler("gauge", storage))
	mux.Handle(`/update/`, http.HandlerFunc(unknowMetricHandler))
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}

func unknowMetricHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "unknown metric type", http.StatusBadRequest)
}

func UpdateMetricHandler(metricType string, storage repository.MetricsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/" + metricType + "/"), "/")
		if len(parts) != 2 {
			http.NotFound(w, r)
			return
		}
		metricName := parts[0]
		metricValue := parts[1]
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

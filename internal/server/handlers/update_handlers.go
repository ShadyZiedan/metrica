package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// UpdateMetricHandler handles HTTP requests to update a specific metric.
//
// The function extracts the metric type, name, and value from the request's URL parameters.
// It then calls the repository's FindOrCreate method to find or create a metric with the given name and type.
// If an error occurs during this process, it writes an appropriate HTTP error response and returns.
//
// Depending on the metric type, the function parses the metric value and calls the corresponding repository method:
// - For "counter" type, it parses the value as an int64 and calls UpdateCounter.
// - For "gauge" type, it parses the value as a float64 and calls UpdateGauge.
// If an error occurs during these operations, it writes an appropriate HTTP error response and returns.
//
// If the metric type is neither "counter" nor "gauge", it writes an HTTP error response with a status code of 400 and returns.
func (h *MetricHandler) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	metric, err := h.repository.FindOrCreate(r.Context(), metricName, metricType)
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
		err = h.repository.UpdateCounter(r.Context(), metric.Name, num)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	case "gauge":
		num, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = h.repository.UpdateGauge(r.Context(), metric.Name, num)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	default:
		http.Error(w, "unknown metric type", http.StatusBadRequest)
	}
}

package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

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

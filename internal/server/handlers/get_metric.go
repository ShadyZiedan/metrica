package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *MetricHandler) GetMetric(rw http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metric, err := h.repository.Find(r.Context(), metricName)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}
	switch metricType {
	case "counter":
		io.WriteString(rw, fmt.Sprintf("%v", *metric.Counter))
		return
	case "gauge":
		io.WriteString(rw, fmt.Sprintf("%v", *metric.Gauge))
		return
	default:
		http.Error(rw, "unknown metric type: "+metricType, http.StatusNotFound)
		return
	}
}

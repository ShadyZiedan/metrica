package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// GetMetric retrieves a specific metric based on the provided metric type and name.
// It writes the metric value to the response writer as a string.
//
// The function uses chi.URLParam to extract the metric type and name from the request URL.
// It then calls the Find method of the repository to retrieve the metric.
// If an error occurs during the retrieval, it writes an HTTP error response with the status code 404 (Not Found) and returns.
//
// If the metric type is "counter", it writes the counter value to the response writer as a string.
// If the metric type is "gauge", it writes the gauge value to the response writer as a string.
// If the metric type is neither "counter" nor "gauge", it writes an HTTP error response with the status code 404 (Not Found) and returns.
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

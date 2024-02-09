package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/shadyziedan/metrica/internal/models"
)

func (h *MetricHandler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	data := &models.Metrics{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	metric, err := h.repository.FindOrCreate(data.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch data.MType {
	case "counter":
		err := h.repository.UpdateCounter(metric.Name, *data.Delta)
		if err != nil {
			http.Error(w, fmt.Sprintf("error updating counter metric: %s", err), http.StatusInternalServerError)
			return
		}
	case "gauge":
		err := h.repository.UpdateGauge(metric.Name, *data.Value)
		if err != nil {
			http.Error(w, fmt.Sprintf("error updating counter metric: %s", err), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "unknown metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := &models.Metrics{}
	response.ParseMetricModel(*metric)

	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

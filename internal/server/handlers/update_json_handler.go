package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/shadyziedan/metrica/internal/models"
)

func (h *MetricHandler) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := &models.Metrics{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	metric, err := h.repository.FindOrCreate(ctx, data.ID, data.MType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch data.MType {
	case "counter":
		err := h.repository.UpdateCounter(ctx, metric.Name, *data.Delta)
		if err != nil {
			http.Error(w, fmt.Sprintf("error updating counter metric: %s", err), http.StatusInternalServerError)
			return
		}
	case "gauge":
		err := h.repository.UpdateGauge(ctx, metric.Name, *data.Value)
		if err != nil {
			http.Error(w, fmt.Sprintf("error updating counter metric: %s", err), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "unknown metric type", http.StatusBadRequest)
		return
	}
	updatedMetric, err := h.repository.Find(ctx, metric.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	response := &models.Metrics{}
	response.ParseMetricModel(updatedMetric)

	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

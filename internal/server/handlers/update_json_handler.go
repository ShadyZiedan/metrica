package handlers

import (
	"encoding/json"
	"github.com/shadyziedan/metrica/internal/models"
	"net/http"
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
		metric.MType = data.MType
		metric.UpdateCounter(*data.Delta)
	case "gauge":
		metric.MType = data.MType
		metric.UpdateGauge(*data.Value)
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

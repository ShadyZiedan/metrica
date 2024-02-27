package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/shadyziedan/metrica/internal/models"
)

func (h *MetricHandler) UpdateBatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var data []models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	response := make([]*models.Metrics, 0, len(data))
	for _, item := range data {
		metric, err := h.repository.FindOrCreate(ctx, item.ID, item.MType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		switch item.MType {
		case "counter":
			if item.Delta == nil {
				http.Error(w, fmt.Sprintf("error updating null counter value metric: %s", err), http.StatusBadRequest)
				return
			}
			err := h.repository.UpdateCounter(ctx, metric.Name, *item.Delta)
			if err != nil {
				http.Error(w, fmt.Sprintf("error updating counter metric: %s", err), http.StatusInternalServerError)
				return
			}
		case "gauge":
			if item.Value == nil {
				http.Error(w, fmt.Sprintf("error updating null gauge value metric: %s", err), http.StatusBadRequest)
				return
			}
			err := h.repository.UpdateGauge(ctx, metric.Name, *item.Value)
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
		responseModel := &models.Metrics{}
		responseModel.ParseMetricModel(updatedMetric)
		response = append(response, responseModel)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

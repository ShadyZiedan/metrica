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
	var response []*models.Metrics
	for _, datum := range data {
		metric, err := h.repository.FindOrCreate(ctx, datum.ID, datum.MType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		switch datum.MType {
		case "counter":
			err := h.repository.UpdateCounter(ctx, metric.Name, *datum.Delta)
			if err != nil {
				http.Error(w, fmt.Sprintf("error updating counter metric: %s", err), http.StatusInternalServerError)
				return
			}
		case "gauge":
			err := h.repository.UpdateGauge(ctx, metric.Name, *datum.Value)
			if err != nil {
				http.Error(w, fmt.Sprintf("error updating counter metric: %s", err), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "unknown metric type", http.StatusBadRequest)
			return
		}
		responseModel := &models.Metrics{}
		responseModel.ParseMetricModel(*metric)
		response = append(response, responseModel)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

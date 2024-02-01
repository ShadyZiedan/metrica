package handlers

import (
	"encoding/json"
	"github.com/shadyziedan/metrica/internal/models"
	"net/http"
)

func (h *MetricHandler) GetMetricJson(w http.ResponseWriter, r *http.Request) {
	data := &models.Metrics{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "invalid data format", http.StatusBadRequest)
		return
	}
	metric, err := h.repository.Find(data.ID)
	if err != nil {
		http.Error(w, "metric not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	resp := &models.Metrics{}
	resp.ParseMetricModel(*metric)

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

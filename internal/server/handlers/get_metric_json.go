package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/shadyziedan/metrica/internal/models"
)

// GetMetricJSON retrieves a metric by its ID from the request body, finds it in the repository,
// and returns it as a JSON response.
//
// The function expects a JSON object in the request body with the following structure:
//
//	{
//	  "ID": "string" // The ID of the metric to be retrieved.
//	}
//
// If the request body does not contain a valid JSON object or the ID is missing,
// the function returns a 400 Bad Request status with the message "invalid data format".
//
// If the metric with the given ID is not found in the repository,
// the function returns a 404 Not Found status with the message "metric not found".
//
// The function sets the "Content-Type" header of the response to "application/json".
//
// If an error occurs while encoding the metric to JSON,
// the function returns a 500 Internal Server Error status with the error message.
func (h *MetricHandler) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	data := &models.Metrics{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "invalid data format", http.StatusBadRequest)
		return
	}
	metric, err := h.repository.Find(r.Context(), data.ID)
	if err != nil {
		http.Error(w, "metric not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	resp := &models.Metrics{}
	resp.ParseMetricModel(metric)

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

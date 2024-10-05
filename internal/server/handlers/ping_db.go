package handlers

import "net/http"

// Ping handles a HTTP request to check the database connection.
// It responds with a status code 500 if the database connection is closed.
//
// Return:
// - No return value. If the database connection is closed, it writes an HTTP error response.
func (h *MetricHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if h.conn == nil {
		http.Error(w, "db connection closed", http.StatusInternalServerError)
		return
	}
	if err := h.conn.Ping(r.Context()); err != nil {
		http.Error(w, "db connection closed", http.StatusInternalServerError)
		return
	}
}

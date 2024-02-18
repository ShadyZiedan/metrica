package handlers

import "net/http"

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

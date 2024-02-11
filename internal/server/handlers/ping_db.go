package handlers

import "net/http"

func (h *MetricHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if h.conn == nil || h.conn.IsClosed() {
		http.Error(w, "db connection closed", http.StatusInternalServerError)
		return
	}
}

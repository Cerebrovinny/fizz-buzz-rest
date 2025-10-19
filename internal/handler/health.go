package handler

import "net/http"

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	respondJSON(h.logger, w, http.StatusOK, HealthResponse{Status: "ok", Service: "fizzbuzz-api"})
}

package server

import (
	"encoding/json"
	"net/http"
)

type statsHandler struct {
	server *Server
}

var _ http.Handler = (*statsHandler)(nil)

func NewStatsHandler(s *Server) http.Handler {
	return &statsHandler{server: s}
}

func (h *statsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.server.handleError(w, "Method not allowed", http.StatusMethodNotAllowed, "")
		return
	}
	h.stats(w, r)
}

type statsResponse struct {
	Success       bool `json:"success"`
	ResourceCount int  `json:"resourceCount"`
}

func (h *statsHandler) stats(w http.ResponseWriter, r *http.Request) {
	bucket, err := h.server.bucket()
	if err != nil {
		h.server.handleError(w, "Internal server error", http.StatusInternalServerError, err.Error())
		return
	}

	objects, err := bucket.List()
	if err != nil {
		h.server.handleError(w, "Internal server error", http.StatusInternalServerError, err.Error())
		return
	}

	json.NewEncoder(w).Encode(statsResponse{
		Success:       true,
		ResourceCount: len(objects.Objects),
	})
}

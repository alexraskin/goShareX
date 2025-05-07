package server

import (
	"encoding/json"
	"log"
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
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed\n"))
		return
	}
	h.stats(w, r)
}

type statsResponse struct {
	Success          bool `json:"success"`
	ImageUploadCount int  `json:"imageUploadCount"`
}

func (h *statsHandler) stats(w http.ResponseWriter, r *http.Request) {
	bucket, err := h.server.bucket()
	if err != nil {
		log.Println(err)
		http.Error(w, `{"success": false, "errorMessage": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	objects, err := bucket.List()
	if err != nil {
		log.Println(err)
		http.Error(w, `{"success": false, "errorMessage": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(statsResponse{
		Success:          true,
		ImageUploadCount: len(objects.Objects),
	})
}

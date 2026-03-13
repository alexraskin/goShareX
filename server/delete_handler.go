package server

import (
	"encoding/json"
	"net/http"
)

type deleteResponse struct {
	Success bool `json:"success"`
}

type deleteHandler struct {
	server *Server
}

var _ http.Handler = (*deleteHandler)(nil)

func NewDeleteHandler(s *Server) http.Handler {
	return &deleteHandler{server: s}
}

func (h *deleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.server.handleError(w, "Method not allowed", http.StatusMethodNotAllowed, nil)
		return
	}
	h.delete(w, r)
}

func (h *deleteHandler) delete(w http.ResponseWriter, r *http.Request) {
	if !h.server.authenticate(r) {
		h.server.handleError(w, "Invalid authkey", http.StatusUnauthorized, nil)
		return
	}

	fileName := r.URL.Query().Get("fileName")
	if !validKey(fileName) {
		h.server.handleError(w, "Invalid filename", http.StatusBadRequest, nil)
		return
	}

	bucket, err := h.server.bucket()
	if err != nil {
		h.server.handleError(w, "Internal server error", http.StatusInternalServerError, err)
		return
	}

	err = bucket.Delete(fileName)
	if err != nil {
		h.server.handleError(w, "Internal server error", http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(deleteResponse{Success: true})
}

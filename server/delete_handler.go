package server

import (
	"net/http"
)

type deleteHandler struct {
	server *Server
}

var _ http.Handler = (*deleteHandler)(nil)

func NewDeleteHandler(s *Server) http.Handler {
	return &deleteHandler{server: s}
}

func (h *deleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed\n"))
		return
	}
	h.delete(w, r)
}

func (h *deleteHandler) delete(w http.ResponseWriter, r *http.Request) {
	if !authenticate(r, h.server) {
		h.server.handleError(w, "Invalid authkey", http.StatusUnauthorized, "")
		return
	}

	fileName := r.URL.Query().Get("fileName")
	if fileName == "" {
		h.server.handleError(w, "Missing filename", http.StatusBadRequest, "")
		return
	}

	bucket, err := h.server.bucket()
	if err != nil {
		h.server.handleError(w, "Internal server error", http.StatusInternalServerError, err.Error())
		return
	}

	err = bucket.Delete(fileName)
	if err != nil {
		h.server.handleError(w, "Internal server error", http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

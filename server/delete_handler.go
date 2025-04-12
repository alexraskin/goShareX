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
		http.Error(w, `{"success": false, "message": "Invalid authkey"}`, http.StatusUnauthorized)
		return
	}

	fileName := r.URL.Query().Get("fileName")
	if fileName == "" {
		http.Error(w, `{"success": false, "message": "Missing filename"}`, http.StatusBadRequest)
		return
	}

	bucket, err := h.server.bucket()
	if err != nil {
		handleErr(w, err)
		return
	}

	err = bucket.Delete(fileName)
	if err != nil {
		handleErr(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

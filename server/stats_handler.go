package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
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
		h.server.handleError(w, "Method not allowed", http.StatusMethodNotAllowed, nil)
		return
	}
	h.stats(w, r)
}

type statsResponse struct {
	Success       bool   `json:"success"`
	TinyGoVersion string `json:"tinygoVersion"`
	MemoryUsed    string `json:"memoryUsed"`
	ResourceCount int    `json:"resourceCount"`
	Truncated     bool   `json:"truncated"`
}

func (h *statsHandler) stats(w http.ResponseWriter, r *http.Request) {
	if !authenticate(r, h.server) {
		h.server.handleError(w, "Invalid authkey", http.StatusUnauthorized, nil)
		return
	}

	bucket, err := h.server.bucket()
	if err != nil {
		h.server.handleError(w, "Internal server error", http.StatusInternalServerError, err)
		return
	}

	objects, err := bucket.List()
	if err != nil {
		h.server.handleError(w, "Internal server error", http.StatusInternalServerError, err)
		return
	}
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statsResponse{
		Success:       true,
		TinyGoVersion: runtime.Version(),
		MemoryUsed:    fmt.Sprintf("%s / %s (%s garbage collected)", humanizeBytes(stats.Alloc), humanizeBytes(stats.Sys), humanizeBytes(stats.TotalAlloc)),
		ResourceCount: len(objects.Objects),
		Truncated:     objects.Truncated,
	})
}

func humanizeBytes(bytes uint64) string {
	switch {
	case bytes >= 1<<30:
		return fmt.Sprintf("%.2f GB", float64(bytes)/(1<<30))
	case bytes >= 1<<20:
		return fmt.Sprintf("%.2f MB", float64(bytes)/(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.2f KB", float64(bytes)/(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

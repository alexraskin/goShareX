package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type configHandler struct {
	server *Server
}

var _ http.Handler = (*configHandler)(nil)

func NewConfigHandler(s *Server) http.Handler {
	return &configHandler{server: s}
}

func (h *configHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		h.server.handleError(w, "Method not allowed", http.StatusMethodNotAllowed, nil)
		return
	}
	h.getConfig(w, req)
}

type shareXConfig struct {
	Version         string            `json:"Version"`
	Name            string            `json:"Name"`
	DestinationType string            `json:"DestinationType"`
	RequestMethod   string            `json:"RequestMethod"`
	RequestURL      string            `json:"RequestURL"`
	Headers         map[string]string `json:"Headers"`
	Body            string            `json:"Body"`
	URL             string            `json:"URL"`
	DeletionURL     string            `json:"DeletionURL"`
	ErrorMessage    string            `json:"ErrorMessage"`
}

func (h *configHandler) getConfig(w http.ResponseWriter, req *http.Request) {
	if !h.server.authenticate(req) {
		h.server.handleError(w, "Invalid authkey", http.StatusUnauthorized, nil)
		return
	}
	baseURL := fmt.Sprintf("https://%s", req.Host)

	config := shareXConfig{
		Version:         "14.0.1",
		Name:            "goShareX",
		DestinationType: "ImageUploader, TextUploader, FileUploader",
		RequestMethod:   "POST",
		RequestURL:      baseURL + "/upload",
		Headers: map[string]string{
			"Authorization": "Bearer " + h.server.AuthKey,
		},
		Body:         "Binary",
		URL:          "{json:fileURL}",
		DeletionURL:  "{json:deleteURL}",
		ErrorMessage: "{json:errorMessage}",
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=sharex.sxcu")

	json.NewEncoder(w).Encode(config)
}

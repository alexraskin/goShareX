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
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed\n"))
		return
	}
	h.getConfig(w, req)
}

type shareXConfig struct {
	Version         string     `json:"Version"`
	Name            string     `json:"Name"`
	DestinationType string     `json:"DestinationType"`
	RequestMethod   string     `json:"RequestMethod"`
	RequestURL      string     `json:"RequestURL"`
	Parameters      parameters `json:"Parameters"`
	Body            string     `json:"Body"`
	FileFormName    string     `json:"FileFormName"`
	URL             string     `json:"URL"`
	DeletionURL     string     `json:"DeletionURL"`
	ErrorMessage    string     `json:"ErrorMessage"`
}

type parameters struct {
	AuthKey string `json:"authKey"`
}

func (h *configHandler) getConfig(w http.ResponseWriter, req *http.Request) {
	if !authenticate(req, h.server) {
		h.server.handleError(w, "Invalid authkey", http.StatusUnauthorized, "")
		return
	}
	baseURL := fmt.Sprintf("https://%s", req.Host)

	config := shareXConfig{
		Version:         "14.0.1",
		Name:            "Sadge Uploader",
		DestinationType: "ImageUploader, TextUploader, FileUploader",
		RequestMethod:   "POST",
		RequestURL:      baseURL + "/upload",
		Parameters: parameters{
			AuthKey: h.server.AuthKey,
		},
		Body:         "Binary",
		FileFormName: "file",
		URL:          "{json:fileURL}",
		DeletionURL:  "{json:deleteURL}",
		ErrorMessage: "{json:errorMessage}",
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=sharex.sxcu")

	json.NewEncoder(w).Encode(config)
}

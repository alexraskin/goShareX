package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/syumai/workers/cloudflare/r2"
)

type Server struct {
	AuthKey    string
	BucketName string

	mux     *http.ServeMux
	muxOnce sync.Once
}

func (s *Server) bucket() (*r2.Bucket, error) {
	return r2.NewBucket(s.BucketName)
}

func (s *Server) initMux() {
	s.mux = http.NewServeMux()

	s.mux.Handle("/upload", NewUploadHandler(s))
	s.mux.Handle("/delete", NewDeleteHandler(s))
	s.mux.Handle("/config", NewConfigHandler(s))
	s.mux.Handle("/stats", NewStatsHandler(s))

	s.mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			key := strings.TrimPrefix(r.URL.Path, "/")
			s.getKey(w, r, key)
			return
		}
		s.handleError(w, "Not found", http.StatusNotFound, "")
	}))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.muxOnce.Do(s.initMux)
	s.mux.ServeHTTP(w, r)
}

func authenticate(req *http.Request, s *Server) bool {
	authKey := req.URL.Query().Get("authKey")
	return authKey == s.AuthKey
}

type errorResponse struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"errorMessage"`
}

func (s *Server) handleError(w http.ResponseWriter, message string, status int, errDetail string) {
	fullMsg := message
	if errDetail != "" {
		fullMsg = message + ": " + errDetail
	}
	response := errorResponse{
		Success:      false,
		ErrorMessage: fullMsg,
	}
	log.Println(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

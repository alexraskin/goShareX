package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/syumai/workers/cloudflare/r2"
)

type Server struct {
	AuthKey    string
	BucketName string
}

func (s *Server) bucket() (*r2.Bucket, error) {
	return r2.NewBucket(s.BucketName)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()

	mux.Handle("/upload", NewUploadHandler(s))
	mux.Handle("/delete", NewDeleteHandler(s))
	mux.Handle("/config", NewConfigHandler(s))
	mux.Handle("/stats", NewStatsHandler(s))

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			key := strings.TrimPrefix(r.URL.Path, "/")
			s.getKey(w, r, key)
			return
		}
		s.handleError(w, "Not found", http.StatusNotFound, "")
	}))

	mux.ServeHTTP(w, r)
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

package server

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/syumai/workers/cloudflare/r2"
)

var _ http.Handler = (*Server)(nil)

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
		s.handleError(w, "Not found", http.StatusNotFound, nil)
	}))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.muxOnce.Do(s.initMux)
	s.mux.ServeHTTP(w, r)
}

func (s *Server) authenticate(req *http.Request) bool {
	authKey := req.Header.Get("Authorization")
	authKey = strings.TrimPrefix(authKey, "Bearer ")
	if len(authKey) == 0 {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(authKey), []byte(s.AuthKey)) == 1
}

func validKey(key string) bool {
	return len(key) > 0 && len(key) <= 255 &&
		!strings.Contains(key, "/") && !strings.Contains(key, "..")
}

type errorResponse struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"errorMessage"`
}

func (s *Server) handleError(w http.ResponseWriter, message string, status int, err error) {
	if err != nil {
		log.Printf("HTTP %d - %s: %v", status, message, err)
	} else {
		log.Printf("HTTP %d - %s", status, message)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{
		Success:      false,
		ErrorMessage: message,
	})
}

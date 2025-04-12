package server

import (
	"net/http"
	"strings"

	"github.com/syumai/workers/cloudflare/r2"
)

type Server struct {
	BucketName string
	AuthKey    string
}

func (s *Server) bucket() (*r2.Bucket, error) {
	return r2.NewBucket(s.BucketName)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()

	mux.Handle("/upload", NewUploadHandler(s))
	mux.Handle("/delete", NewDeleteHandler(s))
	mux.Handle("/config", NewConfigHandler(s))

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			key := strings.TrimPrefix(r.URL.Path, "/")
			s.getKey(w, r, key)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("route not found\n"))
	}))

	mux.ServeHTTP(w, r)
}

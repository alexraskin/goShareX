package server

import (
	"fmt"
	"io"
	"net/http"
)

func (s *Server) getKey(w http.ResponseWriter, r *http.Request, key string) {
	bucket, err := s.bucket()
	if err != nil {
		s.handleError(w, "Internal server error", http.StatusInternalServerError, err.Error())
		return
	}
	resource, err := bucket.Get(key)
	if err != nil {
		s.handleError(w, "Internal server error", http.StatusInternalServerError, err.Error())
		return
	}
	if resource == nil {
		s.handleError(w, fmt.Sprintf("resource not found: %s", key), http.StatusNotFound, "")
		return
	}
	w.Header().Set("ETag", fmt.Sprintf("W/%s", resource.HTTPETag))
	contentType := "application/octet-stream"
	if resource.HTTPMetadata.ContentType != "" {
		contentType = resource.HTTPMetadata.ContentType
	}
	w.Header().Set("Content-Type", contentType)
	io.Copy(w, resource.Body)
}

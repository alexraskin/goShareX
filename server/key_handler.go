package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func (s *Server) getKey(w http.ResponseWriter, r *http.Request, key string) {
	bucket, err := s.bucket()
	if err != nil {
		log.Println(err)
		http.Error(w, `{"success": false, "errorMessage": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	resource, err := bucket.Get(key)
	if err != nil {
		log.Println(err)
		http.Error(w, `{"success": false, "errorMessage": "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	if resource == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("resource not found: %s", key)))
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

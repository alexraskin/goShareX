package server

import (
	"fmt"
	"io"
	"net/http"
)

func (s *Server) getKey(w http.ResponseWriter, r *http.Request, key string) {
	bucket, err := s.bucket()
	if err != nil {
		handleErr(w, err)
		return
	}
	imgObj, err := bucket.Get(key)
	if err != nil {
		handleErr(w, err)
		return
	}
	if imgObj == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("image not found: %s", key)))
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=14400")
	w.Header().Set("ETag", fmt.Sprintf("W/%s", imgObj.HTTPETag))
	contentType := "application/octet-stream"
	if imgObj.HTTPMetadata.ContentType != "" {
		contentType = imgObj.HTTPMetadata.ContentType
	}
	w.Header().Set("Content-Type", contentType)
	io.Copy(w, imgObj.Body)
}

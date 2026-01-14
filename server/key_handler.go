package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/syumai/workers/cloudflare"
	"github.com/syumai/workers/cloudflare/cache"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	rw.body = append(rw.body, data...)
	return rw.ResponseWriter.Write(data)
}

func (rw *responseWriter) toHTTPResponse() *http.Response {
	return &http.Response{
		StatusCode: rw.statusCode,
		Header:     rw.Header(),
		Body:       io.NopCloser(bytes.NewReader(rw.body)),
	}
}

func (s *Server) getKey(w http.ResponseWriter, r *http.Request, key string) {
	c := cache.New()

	r.Body = nil

	cachedRes, err := c.Match(r, nil)
	if err == nil && cachedRes != nil {
		for k, v := range cachedRes.Header {
			for _, val := range v {
				w.Header().Add(k, val)
			}
		}
		w.Header().Set("X-Cache", "HIT")
		w.WriteHeader(cachedRes.StatusCode)
		io.Copy(w, cachedRes.Body)
		return
	}

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

	contentType := "application/octet-stream"
	if resource.HTTPMetadata.ContentType != "" {
		contentType = resource.HTTPMetadata.ContentType
	}

	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	rw.Header().Set("ETag", fmt.Sprintf("W/%s", resource.HTTPETag))
	rw.Header().Set("Content-Type", contentType)
	rw.Header().Set("Cache-Control", "public, max-age=604800")
	rw.Header().Set("X-Cache", "MISS")

	io.Copy(rw, resource.Body)

	cloudflare.WaitUntil(func() {
		if err := c.Put(r, rw.toHTTPResponse()); err != nil {
			log.Printf("cache put error: %v", err)
		}
	})
}

package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/syumai/workers/cloudflare/r2"
)

type uploadHandler struct {
	server *Server
}

var _ http.Handler = (*uploadHandler)(nil)

func NewUploadHandler(s *Server) http.Handler {
	return &uploadHandler{server: s}
}

func (h *uploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed\n"))
		return
	}
	h.upload(w, r)
}

func (h *uploadHandler) upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fileSlug := nanoID(6)

	contentType := r.Header.Get("Content-Type")
	contentLength := r.Header.Get("Content-Length")

	if contentType == "" || contentLength == "" {
		http.Error(w, `{"success": false, "message": "Missing content-length or content-type"}`, http.StatusBadRequest)
		return
	}

	extensionMap := map[string]string{
		"image/jpeg":    "jpg",
		"image/png":     "png",
		"image/gif":     "gif",
		"image/webp":    "webp",
		"image/svg+xml": "svg",
	}
	fileExt, ok := extensionMap[contentType]
	fileName := fileSlug
	if ok {
		fileName += "." + fileExt
	}

	bucket, err := h.server.bucket()
	if err != nil {
		handleErr(w, err)
		return
	}

	_, err = bucket.Put(fileName, r.Body, &r2.PutOptions{
		HTTPMetadata: r2.HTTPMetadata{
			ContentType: contentType,
		},
	})
	if err != nil {
		handleErr(w, err)
		return
	}

	baseURL := fmt.Sprintf("%s://%s", getScheme(r), r.Host)

	imageURL := fmt.Sprintf("%s/%s", baseURL, fileName)
	deleteURL := fmt.Sprintf("%s/delete?authKey=%s&fileName=%s", baseURL, h.server.AuthKey, fileName)

	resp := map[string]interface{}{
		"success":   true,
		"image":     imageURL,
		"deleteUrl": deleteURL,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)

}

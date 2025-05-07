package server

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/syumai/workers/cloudflare/r2"
)

var extensionMap = map[string]string{
	"image/jpeg":      "jpg",
	"image/png":       "png",
	"image/gif":       "gif",
	"image/webp":      "webp",
	"image/svg+xml":   "svg",
	"image/heic":      "heic",
	"image/heif":      "heif",
	"video/mp4":       "mp4",
	"video/webm":      "webm",
	"video/mov":       "mov",
	"video/avi":       "avi",
	"video/mkv":       "mkv",
	"video/flv":       "flv",
	"text/plain":      "txt",
	"application/pdf": "pdf",
}

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

type uploadHandler struct {
	server *Server
}

var _ http.Handler = (*uploadHandler)(nil)

func NewUploadHandler(s *Server) http.Handler {
	return &uploadHandler{server: s}
}

func (h *uploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.server.handleError(w, "Method not allowed", http.StatusMethodNotAllowed, "")
		return
	}
	h.upload(w, r)
}

type uploadResponse struct {
	Success bool   `json:"success"`
	File    string `json:"fileURL"`
	Delete  string `json:"deleteURL"`
	Error   string `json:"errorMessage"`
}

func (h *uploadHandler) upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.server.handleError(w, "Method not allowed", http.StatusMethodNotAllowed, "")
		return
	}
	fileSlug := randomID(6)

	contentType := r.Header.Get("Content-Type")
	contentLength := r.Header.Get("Content-Length")

	if contentType == "" || contentLength == "" {
		h.server.handleError(w, "Missing content-length or content-type", http.StatusBadRequest, "")
		return
	}

	fileExt, ok := extensionMap[contentType]
	fileName := fileSlug
	if ok {
		fileName += "." + fileExt
	}

	bucket, err := h.server.bucket()
	if err != nil {
		h.server.handleError(w, "Internal server error", http.StatusInternalServerError, err.Error())
		return
	}

	objects, err := bucket.List()
	if err != nil {
		h.server.handleError(w, "Internal server error", http.StatusInternalServerError, err.Error())
		return
	}

	for _, obj := range objects.Objects {
		if obj.Key == fileName {
			h.server.handleError(w, "File already exists", http.StatusBadRequest, "")
			return
		}
	}

	_, err = bucket.Put(fileName, r.Body, &r2.PutOptions{
		HTTPMetadata: r2.HTTPMetadata{
			ContentType:  contentType,
			CacheControl: "public, max-age=604800",
		},
		CustomMetadata: map[string]string{
			"filename": fileName,
			"date":     time.Now().Format(time.RFC3339),
		},
	})
	if err != nil {
		h.server.handleError(w, "Internal server error", http.StatusInternalServerError, err.Error())
		return
	}

	baseURL := fmt.Sprintf("https://%s", r.Host)

	resourceURL := fmt.Sprintf("%s/%s", baseURL, fileName)
	deleteURL := fmt.Sprintf("%s/delete?authKey=%s&fileName=%s", baseURL, h.server.AuthKey, fileName)

	resp := uploadResponse{
		Success: true,
		File:    resourceURL,
		Delete:  deleteURL,
		Error:   "",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)

}

func randomID(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var id strings.Builder
	for i := 0; i < length; i++ {
		id.WriteByte(letters[seededRand.Intn(len(letters))])
	}
	return id.String()
}

package server

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strconv"
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

type uploadHandler struct {
	server *Server
}

var _ http.Handler = (*uploadHandler)(nil)

func NewUploadHandler(s *Server) http.Handler {
	return &uploadHandler{server: s}
}

func (h *uploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.server.handleError(w, "Method not allowed", http.StatusMethodNotAllowed, nil)
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
	if !authenticate(r, h.server) {
		h.server.handleError(w, "Invalid authkey", http.StatusUnauthorized, nil)
		return
	}

	contentType := r.Header.Get("Content-Type")

	if contentType == "" || r.ContentLength == 0 {
		h.server.handleError(w, "Missing content-type or content-length", http.StatusBadRequest, nil)
		return
	}

	ext, ok := extensionMap[contentType]

	bucket, err := h.server.bucket()
	if err != nil {
		h.server.handleError(w, "Internal server error", http.StatusInternalServerError, err)
		return
	}

	const maxRetries = 5
	var fileName string
	for i := 0; i < maxRetries; i++ {
		fileName = randomID(6)
		if ok {
			fileName += "." + ext
		}
		existing, err := bucket.Head(fileName)
		if err != nil {
			h.server.handleError(w, "Internal server error", http.StatusInternalServerError, err)
			return
		}
		if existing == nil {
			break
		}
		if i == maxRetries-1 {
			h.server.handleError(w, "Failed to generate unique file ID", http.StatusInternalServerError, nil)
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
			"size":     strconv.FormatInt(r.ContentLength, 10),
		},
	})
	if err != nil {
		h.server.handleError(w, "Internal server error", http.StatusInternalServerError, err)
		return
	}

	baseURL := fmt.Sprintf("https://%s", r.Host)

	resourceURL := fmt.Sprintf("%s/%s", baseURL, fileName)
	deleteURL := fmt.Sprintf("%s/delete?fileName=%s", baseURL, fileName)

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
		id.WriteByte(letters[rand.IntN(len(letters))])
	}
	return id.String()
}

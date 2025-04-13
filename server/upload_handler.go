package server

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/syumai/workers/cloudflare/r2"
)

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
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed\n"))
		return
	}
	h.upload(w, r)
}

type uploadResponse struct {
	Success bool   `json:"success"`
	Image   string `json:"image"`
	Delete  string `json:"delete"`
}

func (h *uploadHandler) upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fileSlug := randomID(6)

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
		log.Println(err)
		http.Error(w, `{"success": false, "message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	log.Println(fileName)

	objects, err := bucket.List()
	if err != nil {
		log.Println(err)
		http.Error(w, `{"success": false, "message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	for _, obj := range objects.Objects {
		log.Println(obj.Key)
		if obj.Key == fileName {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, `{"success": false, "message": "File already exists"}`, http.StatusBadRequest)
			return
		}
	}

	_, err = bucket.Put(fileName, r.Body, &r2.PutOptions{
		HTTPMetadata: r2.HTTPMetadata{
			ContentType: contentType,
		},
		CustomMetadata: map[string]string{
			"filename": fileName,
			"date":     time.Now().Format(time.RFC3339),
		},
	})
	if err != nil {
		log.Println(err)
		http.Error(w, `{"success": false, "message": "Internal server error"}`, http.StatusInternalServerError)
		return
	}

	baseURL := fmt.Sprintf("https://%s", r.Host)

	imageURL := fmt.Sprintf("%s/%s", baseURL, fileName)
	deleteURL := fmt.Sprintf("%s/delete?authKey=%s&fileName=%s", baseURL, h.server.AuthKey, fileName)

	resp := uploadResponse{
		Success: true,
		Image:   imageURL,
		Delete:  deleteURL,
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

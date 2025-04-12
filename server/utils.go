package server

import (
	"log"
	"net/http"
	"strings"
)

var seed uint32 = 12345

func rand() uint32 {
	seed = (seed*1103515245 + 12345) & 0x7fffffff
	return seed
}

func handleErr(w http.ResponseWriter, err error) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": false, "message": "Internal Server Error"}`))
}

func nanoID(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var id strings.Builder
	for i := 0; i < length; i++ {
		id.WriteByte(letters[rand()%uint32(len(letters))])
	}
	return id.String()
}

func getScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func authenticate(req *http.Request, s *Server) bool {
	authKey := req.URL.Query().Get("authKey")
	return authKey == s.AuthKey
}

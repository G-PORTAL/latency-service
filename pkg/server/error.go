package server

import (
	"net/http"
)

// handleError Handle server error
func handleError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte("Internal server error"))
}

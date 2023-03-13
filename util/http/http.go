package http

import (
	"encoding/json"
	"net/http"
)

// JSON replies to the request with the specified JSON data and HTTP code.
func JSON(w http.ResponseWriter, v interface{}, code int) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}

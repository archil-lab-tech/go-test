package common

import (
	"encoding/json"
	"net/http"
)

// WriteJSON writes a JSON response with status code and sets the content type.
func WriteJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// internal/handlers/debug.go
package handlers

import (
	"encoding/json"
	"net/http"
	"os"
)

func EnvProbe(w http.ResponseWriter, r *http.Request) {
	_, has := os.LookupEnv("MONGO_URI")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":            true,
		"has_MONGO_URI": has,
	})
}

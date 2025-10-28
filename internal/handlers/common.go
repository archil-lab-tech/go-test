package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"archil.lab.tech.com/internal/config"
)

var runtime *config.Config

func SetRuntime(cfg *config.Config) { runtime = cfg }

type OK struct {
	OK      bool   `json:"ok"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Time    string `json:"time"`
}

type ERR struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
	Code  int    `json:"code"`
	Time  string `json:"time"`
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

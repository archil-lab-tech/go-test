package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"archil.lab.tech.com/internal/common"
	"go.mongodb.org/mongo-driver/bson"
)

func Status(w http.ResponseWriter, r *http.Request) {
	type resp struct {
		OK   bool           `json:"ok"`
		Data map[string]any `json:"data"`
		Time time.Time      `json:"time"`
	}
	out := resp{OK: true, Data: map[string]any{"db": "app"}, Time: time.Now().UTC()}

	cfg := common.Runtime()
	if cfg == nil {
		out.Data["configured"] = false
		out.Data["error"] = "runtime config not set"
		_ = json.NewEncoder(w).Encode(out)
		return
	}

	hasEnv := strings.TrimSpace(cfg.MongoURI) != ""
	out.Data["has_env"] = hasEnv
	out.Data["db"] = cfg.DBName

	if !hasEnv {
		out.Data["configured"] = false
		out.Data["mongo_ok"] = false
		out.Data["error"] = "mongo not configured"
		_ = json.NewEncoder(w).Encode(out)
		return
	}

	connected := cfg.Mongo != nil
	out.Data["configured"] = true
	out.Data["mongo_ok"] = connected

	if connected {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := cfg.DB.RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
			out.Data["mongo_ok"] = false
			out.Data["error"] = "mongo ping failed: " + err.Error()
		}
	} else {
		out.Data["error"] = "mongo not connected (likely Atlas allowlist)"
	}

	_ = json.NewEncoder(w).Encode(out)
}

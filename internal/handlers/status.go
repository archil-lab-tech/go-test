package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func Status(w http.ResponseWriter, r *http.Request) {
	type resp struct {
		OK   bool           `json:"ok"`
		Data map[string]any `json:"data"`
		Time time.Time      `json:"time"`
	}
	out := resp{OK: true, Data: map[string]any{"db": "app"}, Time: time.Now().UTC()}

	if runtimeCfg == nil {
		out.Data["configured"] = false
		out.Data["error"] = "runtime config not set"
		_ = json.NewEncoder(w).Encode(out)
		return
	}

	hasEnv := strings.TrimSpace(runtimeCfg.MongoURI) != ""
	out.Data["has_env"] = hasEnv
	out.Data["db"] = runtimeCfg.DBName

	if !hasEnv {
		out.Data["configured"] = false
		out.Data["mongo_ok"] = false
		out.Data["error"] = "mongo not configured"
		_ = json.NewEncoder(w).Encode(out)
		return
	}

	// We have an env; check whether the client is connected.
	connected := runtimeCfg.Mongo != nil
	out.Data["configured"] = true
	out.Data["mongo_ok"] = connected

	// If connected, do a quick ping to be sure.
	if connected {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := runtimeCfg.DB.RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
			out.Data["mongo_ok"] = false
			out.Data["error"] = "mongo ping failed: " + err.Error()
		}
	} else {
		out.Data["error"] = "mongo not connected (very likely blocked by Atlas IP allowlist)"
	}

	_ = json.NewEncoder(w).Encode(out)
}

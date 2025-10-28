package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"archil.lab.tech.com/internal/common"
	"archil.lab.tech.com/internal/config"
	"go.mongodb.org/mongo-driver/bson"
)

func Status(w http.ResponseWriter, r *http.Request) {
	type resp struct {
		OK   bool           `json:"ok"`
		Data map[string]any `json:"data"`
		Time time.Time      `json:"time"`
	}
	out := resp{OK: true, Data: map[string]any{}, Time: time.Now().UTC()}

	cfg := common.Runtime()
	if cfg == nil {
		out.Data["configured"] = false
		out.Data["error"] = "runtime config not set"
		_ = json.NewEncoder(w).Encode(out)
		return
	}

	out.Data["db"] = cfg.DBName
	out.Data["has_env"] = strings.TrimSpace(cfg.MongoURI) != ""

	// If we have a URI but no client (e.g., early boot/rotated secret), try to connect now.
	if out.Data["has_env"].(bool) && cfg.Mongo == nil {
		config.EnsureMongo(r.Context(), cfg)
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
			// Optionally clear so next call re-attempts a fresh connect:
			// config.CloseMongo(r.Context(), cfg)
		}
	} else if out.Data["has_env"].(bool) {
		out.Data["error"] = "mongo not connected (will retry)"
	} else {
		out.Data["error"] = "mongo not configured"
	}

	_ = json.NewEncoder(w).Encode(out)
}

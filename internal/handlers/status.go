package handlers

import (
	"context"
	"net/http"
	"time"
)

// Status godoc
// @Summary Mongo status
// @Produce json
// @Success 200 {object} OK
// @Router /status [get]
func Status(w http.ResponseWriter, r *http.Request) {
	res := map[string]any{
		"configured": runtime != nil && runtime.Mongo != nil,
		"db":         runtime.DBName,
	}
	if runtime != nil && runtime.Mongo != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		err := runtime.Mongo.Ping(ctx, nil)
		res["mongo_ok"] = (err == nil)
		if err != nil {
			res["error"] = err.Error()
		}
	} else {
		res["mongo_ok"] = false
		res["error"] = "mongo not configured"
	}
	writeJSON(w, http.StatusOK, OK{OK: true, Data: res, Time: now()})
}

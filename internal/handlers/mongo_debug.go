package handlers

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"archil.lab.tech.com/internal/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

func MongoDiag(w http.ResponseWriter, r *http.Request) {
	type dnsRes struct {
		Host string   `json:"host"`
		IPs  []string `json:"ips,omitempty"`
		Err  string   `json:"err,omitempty"`
	}
	type out struct {
		OK        bool     `json:"ok"`
		HasEnv    bool     `json:"has_env"`
		Connected bool     `json:"connected"`
		Error     string   `json:"error,omitempty"`
		Hosts     []dnsRes `json:"hosts,omitempty"`
	}

	cfg := common.Runtime()
	resp := out{OK: true}
	if cfg == nil {
		resp.Error = "runtime config not set"
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	uri := strings.TrimSpace(cfg.MongoURI)
	resp.HasEnv = uri != ""
	if uri == "" {
		resp.Error = "MONGO_URI empty"
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	// List DNS resolutions for hosts in the connection string
	if cs, err := connstring.ParseAndValidate(uri); err == nil {
		for _, raw := range cs.Hosts {
			host := raw
			// If a :port suffix is present, strip it for LookupHost
			if h, _, e := net.SplitHostPort(raw); e == nil && h != "" {
				host = h
			} else if idx := strings.Index(raw, ":"); idx > 0 {
				host = raw[:idx]
			}
			entry := dnsRes{Host: raw}
			ips, err := net.DefaultResolver.LookupHost(context.Background(), host)
			if err != nil {
				entry.Err = err.Error()
			} else {
				entry.IPs = ips
			}
			resp.Hosts = append(resp.Hosts, entry)
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	clOpts := options.Client().
		ApplyURI(uri).
		SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)).
		SetAppName("go-api-diag")
	cl, err := mongo.Connect(ctx, clOpts)
	if err != nil {
		resp.Error = "connect: " + err.Error()
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	defer func() { _ = cl.Disconnect(context.Background()) }()

	pctx, pcancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer pcancel()
	if err := cl.Ping(pctx, nil); err != nil {
		resp.Error = "ping: " + err.Error()
	} else {
		resp.Connected = true
	}
	_ = json.NewEncoder(w).Encode(resp)
}

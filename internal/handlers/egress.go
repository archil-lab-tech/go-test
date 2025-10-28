package handlers

import (
	"io"
	"net/http"
)

// EgressIP returns the public egress IP of this service.
// Use /egress-ip        -> ifconfig.me
//
//	/egress-ip?alt=1  -> ipv4.icanhazip.com (fallback)
func EgressIP(w http.ResponseWriter, r *http.Request) {
	target := "https://ifconfig.me"
	if r.URL.Query().Has("alt") {
		target = "https://ipv4.icanhazip.com"
	}

	resp, err := http.Get(target)
	if err != nil {
		http.Error(w, "egress check failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write(b)
}

package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // same-origin in Cloud Run, OK here
}

// UI simple HTML page
func UI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!doctype html>
<html><head><meta charset="utf-8"/><title>WS Demo</title>
<style>body{font-family:system-ui,Arial;margin:2rem} pre{background:#111;color:#0f0;padding:1rem;border-radius:8px}</style>
</head><body>
<h1>WebSocket Demo</h1>
<p>Streaming server time & Mongo status every second:</p>
<pre id="out"></pre>
<script>
const out = document.getElementById('out');
const proto = location.protocol === 'https:' ? 'wss' : 'ws';
const url = proto + '://' + location.host + '/ws';
const ws = new WebSocket(url);
ws.onmessage = (e) => { out.textContent += e.data + "\n"; };
ws.onclose = () => { out.textContent += "\\n[closed]\\n"; };
</script>
</body></html>`))
}

// WS sends JSON lines {time, mongo_ok}
func WS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx := r.Context()
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			status := map[string]any{"time": now()}
			if runtime != nil && runtime.Mongo != nil {
				c, cancel := context.WithTimeout(ctx, 1*time.Second)
				defer cancel()
				status["mongo_ok"] = (runtime.Mongo.Ping(c, nil) == nil)
			} else {
				status["mongo_ok"] = false
			}
			_ = conn.WriteJSON(status)
		}
	}
}

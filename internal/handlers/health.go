package handlers

import "net/http"

// Root godoc
// @Summary Root JSON
// @Produce json
// @Success 200 {object} OK
// @Router / [get]
func RootJSON(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, OK{OK: true, Message: "ok", Time: now()})
}

// Healthz godoc
// @Summary Liveness
// @Produce json
// @Success 200 {object} OK
// @Router /healthz [get]
func Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, OK{OK: true, Message: "alive", Time: now()})
}

// Readyz godoc
// @Summary Readiness (extend with checks)
// @Produce json
// @Success 200 {object} OK
// @Router /readyz [get]
func Readyz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, OK{OK: true, Message: "ready", Time: now()})
}

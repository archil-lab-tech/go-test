package mid

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &wrap{ResponseWriter: w, status: 200}
		next.ServeHTTP(ww, r)
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", ww.status).
			Dur("dur", time.Since(start)).
			Msg("http")
	})
}

type wrap struct {
	http.ResponseWriter
	status int
}

func (w *wrap) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

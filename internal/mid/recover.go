package mid

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

func RecoverJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error().Interface("panic", rec).Msg("panic recovered")
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				http.Error(w, `{"ok":false,"error":"internal","code":500}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

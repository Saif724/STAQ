package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

func Logger(logg zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()

				next.ServeHTTP(w, r)

				logg.Info().
					Str("module", "http").
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Dur("duration", time.Since(start)).
					Msg("HTTP request")
			},
		)
	}
}

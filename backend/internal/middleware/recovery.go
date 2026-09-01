package middleware

import (
	"net/http"

	"github.com/Saif724/STAQ/backend/internal/shared/response"
	"github.com/rs/zerolog"
)

func Recovery(logg zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logg.Error().
						Str("module", "http").
						Str("request_id", GetRequestID(r.Context())).
						Interface("panic", err).
						Msg("Panic recovered")

					response.ErrorJSON(
						w,
						http.StatusInternalServerError,
						"INTERNAL_ERROR",
						"Internal server error",
					)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
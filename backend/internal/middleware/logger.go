package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int){
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(body []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}

	return rw.ResponseWriter.Write(body)
}

func Logger(logg zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()

				rw := &responseWriter{
					ResponseWriter: w,
				}

				requestID := GetRequestID(r.Context())

				next.ServeHTTP(rw, r)

				logg.Info().
					Str("module", "http").
					Str("request_id", requestID).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Int("status", rw.statusCode).
					Dur("duration", time.Since(start)).
					Msg("HTTP request")
			},
		)
	}
}

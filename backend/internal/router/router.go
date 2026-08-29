package router

import (
	"net/http"

	"github.com/Saif724/STAQ/backend/internal/health"
	"github.com/Saif724/STAQ/backend/internal/middleware"
	"github.com/rs/zerolog"
)

func New(healthHandler *health.Handler,logg zerolog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler.Check)

	handler := middleware.Logger(logg)(mux)

	return handler
}
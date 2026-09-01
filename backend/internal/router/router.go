package router

import (
	"net/http"

	"github.com/Saif724/STAQ/backend/internal/health"
	"github.com/Saif724/STAQ/backend/internal/middleware"
	"github.com/rs/zerolog"
)

func New(
	healthHandler *health.Handler,
	logg zerolog.Logger,
	frontendURL string,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler.Check)

	handler := middleware.Recovery(logg)(
		middleware.CORS(frontendURL)(
			middleware.RequestID(
				middleware.Logger(logg)(mux),
			),
		),
	)

	return handler
}
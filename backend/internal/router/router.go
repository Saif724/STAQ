package router

import (
	"net/http"

	"github.com/Saif724/STAQ/backend/internal/auth"
	"github.com/Saif724/STAQ/backend/internal/health"
	"github.com/Saif724/STAQ/backend/internal/middleware"
	"github.com/Saif724/STAQ/backend/internal/users"
	"github.com/Saif724/STAQ/backend/pkg/jwt"
	"github.com/rs/zerolog"
)

func New(
	healthHandler *health.Handler,
	authHandler *auth.Handler,
	jwtManager *jwt.Manager,
	usersHandler *users.Handler,
	logg zerolog.Logger,
	frontendURL string,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler.Check)
	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.HandleFunc("/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("/auth/logout", authHandler.Logout)
	mux.Handle(
		"/users/me",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(usersHandler.Me),
		),
	)
	mux.HandleFunc("/auth/verify-email", authHandler.VerifyEmail)

	return middleware.Chain(
		mux,
		middleware.Recovery(logg),
		middleware.CORS(frontendURL),
		middleware.RequestID,
		middleware.Logger(logg),
	)
}

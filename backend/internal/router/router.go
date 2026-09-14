package router

import (
	"net/http"

	"github.com/Saif724/STAQ/backend/internal/actions"
	"github.com/Saif724/STAQ/backend/internal/auth"
	"github.com/Saif724/STAQ/backend/internal/health"
	"github.com/Saif724/STAQ/backend/internal/middleware"
	"github.com/Saif724/STAQ/backend/internal/queues"
	"github.com/Saif724/STAQ/backend/internal/tasks"
	"github.com/Saif724/STAQ/backend/internal/triggers"
	"github.com/Saif724/STAQ/backend/internal/users"
	"github.com/Saif724/STAQ/backend/pkg/jwt"
	"github.com/rs/zerolog"
)

func New(
	healthHandler *health.Handler,
	authHandler *auth.Handler,
	jwtManager *jwt.Manager,
	usersHandler *users.Handler,
	taskHandler *tasks.Handler,
	queueHandler *queues.Handler,
	triggersHandler *triggers.Handler,
	actionsHandler *actions.Handler,
	logg zerolog.Logger,
	frontendURL string,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler.Check)

	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /auth/verify-email", authHandler.VerifyEmail)
	mux.HandleFunc("POST /auth/resend-verification", authHandler.ResendVerification)

	mux.HandleFunc("GET /auth/google", authHandler.GoogleLogin)
	mux.HandleFunc("GET /auth/google/callback", authHandler.GoogleCallback)

	mux.Handle(
		"GET /users/me",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(usersHandler.Me),
		),
	)

	mux.Handle(
		"POST /tasks",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(taskHandler.Create),
		),
	)
	mux.Handle(
		"GET /tasks",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(taskHandler.List),
		),
	)
	mux.Handle(
		"GET /tasks/{id}",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(taskHandler.Get),
		),
	)
	mux.Handle(
		"PUT /tasks/{id}",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(taskHandler.Update),
		),
	)
	mux.Handle(
		"DELETE /tasks/{id}",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(taskHandler.Delete),
		),
	)

	mux.Handle(
		"GET /queues",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(queueHandler.List),
		),
	)
	mux.Handle(
		"GET /queues/{id}",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(queueHandler.Get),
		),
	)
	mux.Handle(
		"POST /queues",
		http.HandlerFunc(queueHandler.Create),
	)
	mux.Handle(
		"PUT /queues/{id}",
		http.HandlerFunc(queueHandler.Update),
	)
	mux.Handle(
		"DELETE /queues/{id}",
		http.HandlerFunc(queueHandler.Delete),
	)

	mux.Handle(
		"POST /triggers",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(triggersHandler.Create),
		),
	)
	mux.Handle(
		"GET /tasks/{taskID}/triggers",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(triggersHandler.ListByTaskID),
		),
	)
	mux.Handle(
		"GET /triggers/{id}",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(triggersHandler.Get),
		),
	)
	mux.Handle(
		"PUT /triggers/{id}",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(triggersHandler.Update),
		),
	)
	mux.Handle(
		"DELETE /triggers/{id}",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(triggersHandler.Delete),
		),
	)

	mux.Handle(
		"POST /actions",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(actionsHandler.Create),
		),
	)
	mux.Handle(
		"GET /tasks/{taskID}/actions",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(actionsHandler.ListByTaskID),
		),
	)
	mux.Handle(
		"GET /actions/{id}",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(actionsHandler.Get),
		),
	)
	mux.Handle(
		"PUT /actions/{id}",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(actionsHandler.Update),
		),
	)
	mux.Handle(
		"DELETE /actions/{id}",
		middleware.Auth(jwtManager)(
			http.HandlerFunc(actionsHandler.Delete),
		),
	)

	return middleware.Chain(
		mux,
		middleware.Recovery(logg),
		middleware.CORS(frontendURL),
		middleware.RequestID,
		middleware.Logger(logg),
	)
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Saif724/STAQ/backend/internal/actions"
	emailAction "github.com/Saif724/STAQ/backend/internal/actions/email"
	httpAction "github.com/Saif724/STAQ/backend/internal/actions/http"
	reminderAction "github.com/Saif724/STAQ/backend/internal/actions/reminder"
	shellAction "github.com/Saif724/STAQ/backend/internal/actions/shell"
	"github.com/Saif724/STAQ/backend/internal/auth"
	"github.com/Saif724/STAQ/backend/internal/broker"
	"github.com/Saif724/STAQ/backend/internal/config"
	"github.com/Saif724/STAQ/backend/internal/connections"
	"github.com/Saif724/STAQ/backend/internal/database"
	"github.com/Saif724/STAQ/backend/internal/health"
	emailIntegration "github.com/Saif724/STAQ/backend/internal/integrations/email"
	"github.com/Saif724/STAQ/backend/internal/logger"
	"github.com/Saif724/STAQ/backend/internal/queues"
	"github.com/Saif724/STAQ/backend/internal/router"
	"github.com/Saif724/STAQ/backend/internal/tasks"
	"github.com/Saif724/STAQ/backend/internal/triggers"
	"github.com/Saif724/STAQ/backend/internal/users"
	"github.com/Saif724/STAQ/backend/pkg/email"
	"github.com/Saif724/STAQ/backend/pkg/jwt"
	"github.com/Saif724/STAQ/backend/pkg/securetoken"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logg := logger.New(logger.Config{
		Environment: cfg.App.Env,
	})

	logg.Info().
		Str("module", "main").
		Msg("Logger initialized successfully")

	db, err := database.NewPostgresPool(cfg.Database.URL)

	if err != nil {
		logg.Fatal().
			Err(err).
			Msg("Failed to initialize PostgreSQL")
	}

	defer db.Close()

	logg.Info().
		Str("module", "database").
		Msg("PostgreSQL connection established")

	redisClient, err := broker.NewRedisClient(
		cfg.Redis.Address,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)

	if err != nil {
		logg.Fatal().
			Err(err).
			Msg("Failed to initialize Redis")
	}

	defer redisClient.Close()

	logg.Info().
		Str("module", "redis").
		Msg("Redis connection established")

	encryptor, err := securetoken.NewEncryptorFromBase64(
		cfg.Encryption.Key,
	)
	if err != nil {
		logg.Fatal().
			Err(err).
			Msg("Failed to initialize token encryptor")
	}

	userRepository := users.NewRepository(db)
	usersService := users.NewService(userRepository)

	jwtManager := jwt.NewManager(cfg.JWT.Secret)
	refreshTokenRepository := auth.NewRefreshTokenRepository(db)
	emailVerificationRepository := auth.NewEmailVerificationRepository(db)
	oauthRepository := auth.NewOAuthAccountRepository(db)

	emailSender := email.NewResendSender(
		cfg.Email.APIKey,
		cfg.Email.From,
	)
	authService := auth.NewService(
		usersService,
		jwtManager,
		refreshTokenRepository,
		emailVerificationRepository,
		emailSender,
	)
	oauthService := auth.NewOAuthService(
		usersService,
		oauthRepository,
		jwtManager,
		refreshTokenRepository,
		cfg.Google,
		redisClient.Client(),
	)
	authHandler := auth.NewHandler(
		authService,
		oauthService,
	)

	gmailOAuthService := emailIntegration.NewGmailOAuthService(
		cfg.Google,
		redisClient.Client(),
	)

	connectionsRepository := connections.NewRepository(db)
	connectionsService := connections.NewService(connectionsRepository)

	gmailIntegrationService := emailIntegration.NewService(
		gmailOAuthService,
		connectionsService,
		encryptor,
	)
	gmailHandler := emailIntegration.NewHandler(gmailIntegrationService)

	usersHandler := users.NewHandler(usersService)

	queuesRepository := queues.NewRepository(db)
	queuesService := queues.NewService(queuesRepository)
	queuesHandler := queues.NewHandler(queuesService)

	tasksRepository := tasks.NewRepository(db)
	tasksService := tasks.NewService(tasksRepository, queuesService)
	tasksHandler := tasks.NewHandler(tasksService)

	triggersRepository := triggers.NewRepository(db)
	triggersService := triggers.NewService(triggersRepository, tasksService)
	triggersHandler := triggers.NewHandler(triggersService)

	actionsRepository := actions.NewRepository(db)
	actionsService := actions.NewService(actionsRepository, tasksService)
	actionsHandler := actions.NewHandler(actionsService)

	actionRegistry := actions.NewRegistry()

	actionRegistry.Register(
		actions.TypeReminder,
		reminderAction.NewExecutor(),
	)

	actionRegistry.Register(
		actions.TypeEmail,
		emailAction.NewExecutor(),
	)

	actionRegistry.Register(
		actions.TypeHTTP,
		httpAction.NewExecutor(&http.Client{
			Timeout: 30 * time.Second,
		}),
	)

	actionRegistry.Register(
		actions.TypeShell,
		shellAction.NewExecutor(
			[]string{},
		),
	)

	healthHandler := health.NewHandler(db, redisClient.Client())

	handler := router.New(
		healthHandler,
		authHandler,
		gmailHandler,
		jwtManager,
		usersHandler,
		tasksHandler,
		queuesHandler,
		triggersHandler,
		actionsHandler,
		logg,
		cfg.App.FrontendURL,
	)

	server := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logg.Info().
		Str("module", "http").
		Str("address", server.Addr).
		Msg("HTTP server starting")

	go func() {
		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			logg.Fatal().
				Err(err).
				Msg("HTTP server failed")
		}
	}()

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-stop

	logg.Info().
		Str("module", "http").
		Msg("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logg.Error().
			Err(err).
			Msg("HTTP server shutdown failed")
	} else {
		logg.Info().
			Str("module", "http").
			Msg("HTTP server shutdown complete")
	}

	logg.Info().
		Str("module", "main").
		Msg("STAQ shutdown complete")
}

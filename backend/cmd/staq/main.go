package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Saif724/STAQ/backend/internal/actions"
	"github.com/Saif724/STAQ/backend/internal/auth"
	"github.com/Saif724/STAQ/backend/internal/broker"
	"github.com/Saif724/STAQ/backend/internal/config"
	"github.com/Saif724/STAQ/backend/internal/connections"
	"github.com/Saif724/STAQ/backend/internal/dashboard"
	"github.com/Saif724/STAQ/backend/internal/database"
	"github.com/Saif724/STAQ/backend/internal/executions"
	"github.com/Saif724/STAQ/backend/internal/health"
	emailIntegration "github.com/Saif724/STAQ/backend/internal/integrations/email"
	"github.com/Saif724/STAQ/backend/internal/logger"
	"github.com/Saif724/STAQ/backend/internal/notifications"
	"github.com/Saif724/STAQ/backend/internal/queues"
	"github.com/Saif724/STAQ/backend/internal/router"
	"github.com/Saif724/STAQ/backend/internal/scheduler"
	"github.com/Saif724/STAQ/backend/internal/tasks"
	"github.com/Saif724/STAQ/backend/internal/tasksetups"
	"github.com/Saif724/STAQ/backend/internal/triggers"
	"github.com/Saif724/STAQ/backend/internal/users"
	"github.com/Saif724/STAQ/backend/internal/workers"
	"github.com/Saif724/STAQ/backend/pkg/email"
	"github.com/Saif724/STAQ/backend/pkg/jwt"
	"github.com/Saif724/STAQ/backend/pkg/securetoken"
	"github.com/rs/zerolog"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logg := logger.New(logger.Config{Environment: cfg.App.Env})

	logg.Info().Str("module", "staq").Msg("STAQ service initializing")

	// ---------- Shared infrastructure ----------

	db, err := database.NewPostgresPool(cfg.Database.URL)
	if err != nil {
		logg.Fatal().Err(err).Msg("Failed to initialize PostgreSQL")
	}
	defer db.Close()
	logg.Info().Str("module", "database").Msg("PostgreSQL connection established")

	redisClient, err := broker.NewRedisClient(
		cfg.Redis.Address,
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Redis.TLS,
	)
	if err != nil {
		logg.Fatal().Err(err).Msg("Failed to initialize Redis")
	}
	defer redisClient.Close()
	logg.Info().Str("module", "redis").Msg("Redis connection established")

	encryptor, err := securetoken.NewEncryptorFromBase64(cfg.Encryption.Key)
	if err != nil {
		logg.Fatal().Err(err).Msg("Failed to initialize token encryptor")
	}

	// ---------- Repositories ----------

	userRepository := users.NewRepository(db)
	tasksRepository := tasks.NewRepository(db)
	triggersRepository := triggers.NewRepository(db)
	actionsRepository := actions.NewRepository(db)
	executionsRepository := executions.NewRepository(db)
	notificationsRepository := notifications.NewRepository(db)
	connectionsRepository := connections.NewRepository(db)
	queuesRepository := queues.NewRepository(db)
	refreshTokenRepository := auth.NewRefreshTokenRepository(db)
	emailVerificationRepository := auth.NewEmailVerificationRepository(db)
	oauthRepository := auth.NewOAuthAccountRepository(db)

	// ---------- Services ----------

	usersService := users.NewService(userRepository)
	queuesService := queues.NewService(queuesRepository)
	tasksService := tasks.NewService(tasksRepository, queuesService)
	triggersService := triggers.NewService(triggersRepository, tasksService)
	actionsService := actions.NewService(actionsRepository, tasksService)
	executionsService := executions.NewService(executionsRepository)
	notificationsService := notifications.NewService(notificationsRepository)
	connectionsService := connections.NewService(connectionsRepository)

	jwtManager := jwt.NewManager(cfg.JWT.Secret)

	emailSender := email.NewGmailSender(
		cfg.Google.ClientID,
		cfg.Google.ClientSecret,
		cfg.Email.GmailRefreshToken,
	)

	gmailOAuthService := emailIntegration.NewGmailOAuthService(
		cfg.Google,
		redisClient.Client(),
	)
	gmailIntegrationService := emailIntegration.NewService(
		gmailOAuthService,
		connectionsService,
		encryptor,
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

	// ---------- Shutdown context ----------

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// ---------- Scheduler ----------

	schedulerService := scheduler.NewService(
		db,
		triggersRepository,
		triggersService,
		redisClient,
		logg,
	)

	// ---------- Worker ----------

	workerRegistry := workers.NewRegistry(
		emailSender,
		gmailIntegrationService,
		notificationsService,
	)
	worker := workers.New(
		redisClient,
		tasksRepository,
		actionsRepository,
		executionsService,
		workerRegistry,
		logg,
	)

	// ---------- HTTP API ----------

	healthHandler := health.NewHandler(db, redisClient.Client())

	authHandler := auth.NewHandler(authService, oauthService, cfg.App.FrontendURL)
	usersHandler := users.NewHandler(usersService)
	dashboardHandler := dashboard.NewHandler(dashboard.NewService(db))
	connectionsHandler := connections.NewHandler(connectionsService)
	gmailHandler := emailIntegration.NewHandler(gmailIntegrationService, cfg.App.FrontendURL)
	queuesHandler := queues.NewHandler(queuesService)
	tasksHandler := tasks.NewHandler(tasksService)
	triggersHandler := triggers.NewHandler(triggersService)
	actionsHandler := actions.NewHandler(actionsService)
	executionsHandler := executions.NewHandler(executionsService)
	notificationsHandler := notifications.NewHandler(notificationsService)
	taskSetupHandler := tasksetups.NewHandler(tasksetups.NewService(
		db,
		tasksService,
		triggersService,
		actionsService,
	))

	handler := router.New(
		healthHandler,
		authHandler,
		gmailHandler,
		connectionsHandler,
		jwtManager,
		usersHandler,
		dashboardHandler,
		tasksHandler,
		taskSetupHandler,
		queuesHandler,
		triggersHandler,
		actionsHandler,
		executionsHandler,
		notificationsHandler,
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

	// ---------- Launch all subsystems ----------

	serverErr := make(chan error, 1)

	go func() {
		logg.Info().Str("module", "http").Str("address", server.Addr).Msg("HTTP server starting")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	var background sync.WaitGroup

	// Set EMBEDDED_BACKGROUND=false to run the API without the scheduler/worker.
	if os.Getenv("EMBEDDED_BACKGROUND") != "false" {
		background.Add(2)

		go func() {
			defer background.Done()
			supervise(ctx, logg, "scheduler", schedulerService.Run)
		}()

		go func() {
			defer background.Done()
			supervise(ctx, logg, "worker", worker.Run)
		}()
	}

	logg.Info().Msg("STAQ is running")

	// ---------- Wait for shutdown ----------

	select {
	case <-ctx.Done():
		logg.Info().Msg("shutdown signal received")
	case err := <-serverErr:
		logg.Error().Err(err).Msg("HTTP server failed; shutting down")
		stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logg.Error().Err(err).Msg("HTTP server shutdown failed")
	}

	done := make(chan struct{})
	go func() {
		background.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(20 * time.Second):
		logg.Warn().Msg("background services did not stop in time")
	}

	logg.Info().Msg("STAQ shutdown complete")
}

// supervise keeps a background component alive: a panic or an error is logged
// and the component restarts after a pause, so it can never take the API down.
func supervise(
	ctx context.Context,
	logg zerolog.Logger,
	name string,
	run func(context.Context) error,
) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logg.Fatal().
						Interface("panic", r).
						Str("component", name).
						Msg("component panicked; exiting for clean restart")
				}
			}()

			if err := run(ctx); err != nil && !errors.Is(err, context.Canceled) {
				logg.Error().Err(err).Str("component", name).Msg("component stopped with error")
			}
		}()

		if ctx.Err() != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(30 * time.Second):
		}
	}
}

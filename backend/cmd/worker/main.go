package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Saif724/STAQ/backend/internal/actions"
	"github.com/Saif724/STAQ/backend/internal/broker"
	"github.com/Saif724/STAQ/backend/internal/config"
	"github.com/Saif724/STAQ/backend/internal/connections"
	"github.com/Saif724/STAQ/backend/internal/database"
	"github.com/Saif724/STAQ/backend/internal/executions"
	emailIntegration "github.com/Saif724/STAQ/backend/internal/integrations/email"
	"github.com/Saif724/STAQ/backend/internal/logger"
	"github.com/Saif724/STAQ/backend/internal/tasks"
	"github.com/Saif724/STAQ/backend/internal/workers"
	"github.com/Saif724/STAQ/backend/pkg/securetoken"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logg := logger.New(logger.Config{
		Environment: cfg.App.Env,
	})

	logg.Info().
		Str("module", "worker").
		Msg("Worker service initializing")

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

	tasksRepository := tasks.NewRepository(db)

	actionRepository := actions.NewRepository(db)
	executionsRepository := executions.NewRepository(db)
	executionsService := executions.NewService(executionsRepository)

	encryptor, err := securetoken.NewEncryptorFromBase64(cfg.Encryption.Key)
	if err != nil {
		logg.Fatal().
			Err(err).
			Msg("Failed to initialize token encryptor")
	}

	connectionsRepository := connections.NewRepository(db)
	connectionsService := connections.NewService(connectionsRepository)

	gmailOAuthService := emailIntegration.NewGmailOAuthService(cfg.Google, redisClient.Client())

	gmailIntegrationService := emailIntegration.NewService(
		gmailOAuthService,
		connectionsService,
		encryptor,
	)

	registry := workers.NewRegistry(gmailIntegrationService)
	worker := workers.New(
		redisClient,
		tasksRepository,
		actionRepository,
		executionsService,
		registry,
		logg,
	)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := worker.Run(ctx); err != nil {
		if err != context.Canceled &&
			err != context.DeadlineExceeded {
			logg.Error().
				Err(err).
				Msg("worker exited with error")
		}
	}

	logg.Info().
		Msg("Worker service shutting down")
}

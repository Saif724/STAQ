package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Saif724/STAQ/backend/internal/broker"
	"github.com/Saif724/STAQ/backend/internal/config"
	"github.com/Saif724/STAQ/backend/internal/database"
	"github.com/Saif724/STAQ/backend/internal/logger"
	"github.com/Saif724/STAQ/backend/internal/queues"
	"github.com/Saif724/STAQ/backend/internal/scheduler"
	"github.com/Saif724/STAQ/backend/internal/tasks"
	"github.com/Saif724/STAQ/backend/internal/triggers"
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
		Str("module", "scheduler").
		Msg("Scheduler service initializing")

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
		cfg.Redis.TLS,
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

	queuesRepository := queues.NewRepository(db)
	queuesService := queues.NewService(queuesRepository)

	tasksRepository := tasks.NewRepository(db)
	tasksService := tasks.NewService(tasksRepository, queuesService)

	triggersRepository := triggers.NewRepository(db)
	triggersService := triggers.NewService(triggersRepository, tasksService)

	schedulerService := scheduler.NewService(
		db,
		triggersRepository,
		triggersService,
		redisClient,
		logg,
	)

	schedulerRunner := scheduler.New(schedulerService, logg)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	healthServer := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	healthServerDone := make(chan struct{})

	go func() {
		defer close(healthServerDone)

		logg.Info().
			Str("address", ":"+port).
			Msg("Scheduler health server starting")

		if err := healthServer.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logg.Error().
				Err(err).
				Msg("Scheduler health server failed")
		}
	}()

	schedulerRunner.Start(ctx)

	<-ctx.Done()

	logg.Info().Msg("Scheduler shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := healthServer.Shutdown(shutdownCtx); err != nil {
		logg.Error().
			Err(err).
			Msg("Scheduler health server shutdown failed")
	}

	<-healthServerDone

	logg.Info().Msg("Scheduler service stopped")
}

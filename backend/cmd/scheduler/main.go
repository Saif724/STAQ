package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

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
	schedulerRunner := scheduler.New(
		schedulerService,
		logg,
	)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	schedulerRunner.Start(ctx)

	<-ctx.Done()

	logg.Info().
		Msg("Scheduler service shutting down")
}

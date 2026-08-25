package main

import (
	"log"

	"github.com/Saif724/STAQ/backend/internal/broker"
	"github.com/Saif724/STAQ/backend/internal/config"
	"github.com/Saif724/STAQ/backend/internal/database"
	"github.com/Saif724/STAQ/backend/internal/logger"
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
}

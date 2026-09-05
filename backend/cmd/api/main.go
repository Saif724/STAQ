package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Saif724/STAQ/backend/internal/auth"
	"github.com/Saif724/STAQ/backend/internal/broker"
	"github.com/Saif724/STAQ/backend/internal/config"
	"github.com/Saif724/STAQ/backend/internal/database"
	"github.com/Saif724/STAQ/backend/internal/health"
	"github.com/Saif724/STAQ/backend/internal/logger"
	"github.com/Saif724/STAQ/backend/internal/router"
	"github.com/Saif724/STAQ/backend/internal/users"
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

	userRepository := users.NewRepository(db)
	usersService := users.NewService(userRepository)

	authService := auth.NewService(usersService)
	authHandler := auth.NewHandler(authService)

	healthHandler := health.NewHandler(db, redisClient)

	handler := router.New(healthHandler, authHandler, logg, cfg.App.FrontendURL)

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
			Str("module", "mail").
			Msg("HTTP server shutdown complete")
	}

	logg.Info().
		Str("module", "main").
		Msg("STAQ shutdown complete")
}

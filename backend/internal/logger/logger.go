package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(cfg Config) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	if cfg.Environment == "development" {
		return zerolog.New(
			zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC3339,
			},
		).With().
			Timestamp().
			Logger()
	}
	return zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger()
}

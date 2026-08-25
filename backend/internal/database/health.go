package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func HealthCheck(ctx context.Context, pool *pgxpool.Pool) error {
	return pool.Ping(ctx)
}

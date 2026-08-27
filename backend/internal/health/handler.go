package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

func NewHandler(db *pgxpool.Pool, redisClient *redis.Client) *Handler {
	return &Handler{
		DB:    db,
		Redis: redisClient,
	}
}

func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	response := map[string]string{
		"status":   "ok",
		"database": "ok",
		"redis":    "ok",
	}

	if err := h.DB.Ping(ctx); err != nil {
		response["status"] = "error"
		response["database"] = "error"
	}

	if err := h.Redis.Ping(ctx).Err(); err != nil {
		response["status"] = "error"
		response["redis"] = "error"
	}

	w.Header().Set("Content-Type", "application/json")

	if response["status"] == "error" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(response)
}

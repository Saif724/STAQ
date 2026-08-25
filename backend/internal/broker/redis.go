package broker

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(address, password string, db int) (*redis.Client, error) {
	if address == "" {
		return nil, fmt.Errorf("redis address is required")
	}

	client := redis.NewClient(&redis.Options{
		Addr: address,
		Password: password,
		DB: db,
		DialTimeout: 5 * time.Second,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}
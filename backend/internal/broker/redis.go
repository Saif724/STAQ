package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedisClient(address, password string, db int) (*Redis, error) {
	if address == "" {
		return nil, fmt.Errorf("redis address is required")
	}

	client := redis.NewClient(&redis.Options{
		Addr:         address,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &Redis{
		client: client,
	}, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}

func (r *Redis) Client() *redis.Client {
	return r.client
}

type TaskJob struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"task_id"`
	TriggerID   string    `json:"trigger_id"`
	ScheduledAt time.Time `json:"scheduled_at"`
	CreatedAt   time.Time `json:"created_at"`
}

const TaskQueue = "staq:tasks"

func (r *Redis) PublishTaskJob(
	ctx context.Context,
	job TaskJob,
) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to encode task job: %w", err)
	}

	if err := r.client.RPush(
		ctx,
		TaskQueue,
		payload,
	).Err(); err != nil {
		return fmt.Errorf("failed to publish task job: %w", err)
	}

	return nil
}

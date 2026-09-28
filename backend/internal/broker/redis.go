package broker

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedisClient(address, password string, db int, useTLS bool) (*Redis, error) {
	if address == "" {
		return nil, fmt.Errorf("redis address is required")
	}

	options := &redis.Options{
		Addr:         address,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	if useTLS {
		options.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	client := redis.NewClient(options)

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

const (
	TaskStream        = "staq:task-stream"
	TaskConsumerGroup = "staq-workers"

	TaskClaimIdleTime = 5 * time.Minute
)

func (r *Redis) EnsureTaskConsumerGroup(ctx context.Context) error {
	err := r.client.XGroupCreateMkStream(
		ctx,
		TaskStream,
		TaskConsumerGroup,
		"0",
	).Err()

	if err == nil {
		return nil
	}

	if isConsumerGroupExistsError(err) {
		return nil
	}

	return fmt.Errorf("failed to create task consumer group: %w", err)
}

func isConsumerGroupExistsError(err error) bool {
	return err != nil &&
		strings.HasPrefix(err.Error(), "BUSYGROUP")
}

func (r *Redis) PublishTaskJob(
	ctx context.Context,
	job TaskJob,
) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to encode task job: %w", err)
	}

	_, err = r.client.XAdd(
		ctx,
		&redis.XAddArgs{
			Stream: TaskStream,
			Values: map[string]interface{}{
				"job": string(payload),
			},
		},
	).Result()

	if err != nil {
		return fmt.Errorf("failed to publish task job: %w", err)
	}

	return nil
}

func (r *Redis) ReadTaskJob(
	ctx context.Context,
	consumer string,
) (string, TaskJob, error) {
	streams, err := r.client.XReadGroup(
		ctx,
		&redis.XReadGroupArgs{
			Group:    TaskConsumerGroup,
			Consumer: consumer,
			Streams:  []string{TaskStream, ">"},
			Count:    1,
			Block:    2 * time.Second,
		},
	).Result()

	if err != nil {
		return "", TaskJob{}, err
	}

	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return "", TaskJob{}, redis.Nil
	}

	message := streams[0].Messages[0]

	rawJob, ok := message.Values["job"].(string)
	if !ok {
		return "", TaskJob{}, fmt.Errorf(
			"task stream message %s has invalid job payload",
			message.ID,
		)
	}

	var job TaskJob

	if err := json.Unmarshal(
		[]byte(rawJob),
		&job,
	); err != nil {
		return "", TaskJob{}, fmt.Errorf(
			"failed to decode task stream message %s: %w",
			message.ID,
			err,
		)
	}

	return message.ID, job, nil
}

func (r *Redis) AcknowledgeTaskJob(
	ctx context.Context,
	messageID string,
) error {
	if messageID == "" {
		return fmt.Errorf("task message id is required")
	}

	if err := r.client.XAck(
		ctx,
		TaskStream,
		TaskConsumerGroup,
		messageID,
	).Err(); err != nil {
		return fmt.Errorf("failed to acknowledge task job: %w", err)
	}

	return nil
}

func (r *Redis) ClaimStaleTaskJobs(
	ctx context.Context,
	consumer string,
) ([]redis.XMessage, error) {
	messages, _, err := r.client.XAutoClaim(
		ctx,
		&redis.XAutoClaimArgs{
			Stream:   TaskStream,
			Group:    TaskConsumerGroup,
			Consumer: consumer,
			MinIdle:  TaskClaimIdleTime,
			Start:    "0-0",
			Count:    10,
		},
	).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to claim stale task jobs: %w", err)
	}

	return messages, nil
}

func (r *Redis) RenewTaskJob(
	ctx context.Context,
	messageID string,
	consumer string,
) error {
	if messageID == "" {
		return fmt.Errorf("task message id is required")
	}

	if consumer == "" {
		return fmt.Errorf("consumer is required")
	}

	_, err := r.client.XClaim(
		ctx,
		&redis.XClaimArgs{
			Stream:   TaskStream,
			Group:    TaskConsumerGroup,
			Consumer: consumer,
			MinIdle:  0,
			Messages: []string{messageID},
		},
	).Result()
	if err != nil {
		return fmt.Errorf("failed to renew task job %s: %w", messageID, err)
	}

	return nil
}

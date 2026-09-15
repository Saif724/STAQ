package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/Saif724/STAQ/backend/internal/broker"
	"github.com/Saif724/STAQ/backend/internal/triggers"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

const (
	DefaultPollInterval = 10 * time.Second
)

type Service struct {
	db                *pgxpool.Pool
	triggerRepository *triggers.Repository
	triggerService    *triggers.Service
	broker            *broker.Redis
	logger            zerolog.Logger

	pollInterval time.Duration
}

func NewService(
	db *pgxpool.Pool,
	triggerRepository *triggers.Repository,
	triggerService *triggers.Service,
	broker *broker.Redis,
	logger zerolog.Logger,
) *Service {
	return &Service{
		db:                db,
		triggerRepository: triggerRepository,
		triggerService:    triggerService,
		broker:            broker,
		logger:            logger,
		pollInterval:      DefaultPollInterval,
	}
}

func (s *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	s.logger.Info().
		Dur("poll_interval", s.pollInterval).
		Msg("scheduler started")

	if err := s.ProcessDueTriggers(ctx); err != nil {
		s.logger.Error().
			Err(err).
			Msg("failed to process due triggers")
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info().Msg("scheduler stopping")
			return ctx.Err()

		case <-ticker.C:
			if err := s.ProcessDueTriggers(ctx); err != nil {
				s.logger.Error().
					Err(err).
					Msg("failed to process due triggers")
			}
		}
	}
}

func (s *Service) ProcessDueTriggers(ctx context.Context) error {
	for {
		processed, err := s.processOne(ctx)

		if err != nil {
			return err
		}

		if !processed {
			return nil
		}
	}
}

func (s *Service) processOne(
	ctx context.Context,
) (bool, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to begin scheduler transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	now := time.Now().UTC()

	trigger, err := s.triggerRepository.FindDueForUpdate(ctx, tx, now)
	if err != nil {
		if err == triggers.ErrTriggerNotFound {
			return false, nil
		}

		return false, fmt.Errorf("failed to fing due trigger: %w", err)
	}

	scheduledTime := trigger.NextRunAt

	job := broker.TaskJob{
		ID:          uuid.NewString(),
		TaskID:      trigger.TaskID,
		ScheduledAt: scheduledTime,
		CreatedAt:   now,
	}

	if err := s.broker.PublishTaskJob(ctx, job); err != nil {
		return false, fmt.Errorf("failed to publish task job: %w", err)
	}

	if trigger.TriggerType == triggers.TypeOnce {
		trigger.LastRunAt = &now
		trigger.IsActive = false
		trigger.UpdatedAt = now

		if err := s.updateTrigger(ctx, tx, trigger); err != nil {
			return false, fmt.Errorf("failed to deactivate one-time trigger: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return false, fmt.Errorf("failed to commit one-time trigger: %w", err)
		}
		s.logger.Info().
			Str("trigger_id", trigger.ID).
			Str("task_id", trigger.TaskID).
			Str("job_id", job.ID).
			Msg("one-time trigger dispatched")

		return true, nil
	}

	nextRunAt, err := s.triggerService.CalculateNextRun(
		trigger,
		scheduledTime,
	)
	if err != nil {
		return false, fmt.Errorf("failed to calculate next run: %w", err)
	}

	trigger.LastRunAt = &now
	trigger.NextRunAt = nextRunAt
	trigger.UpdatedAt = now

	if err := s.updateTrigger(ctx, tx, trigger); err != nil {
		return false, fmt.Errorf("failed to update trigger schedule: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("failed to commit trigger scheduler: %w", err)
	}

	s.logger.Info().
		Str("trigger_id", trigger.ID).
		Str("task_id", trigger.TaskID).
		Str("job_id", job.ID).
		Time("scheduled_at", scheduledTime).
		Time("next_run_at", nextRunAt).
		Msg("trigger dispatched and schedule advanced")

	return true, nil
}

func (s *Service) updateTrigger(
	ctx context.Context,
	tx pgx.Tx,
	trigger *triggers.Trigger,
) error {
	const query = `
		UPDATE triggers
		SET
			next_run_at = $1,
			last_run_at = $2,
			is_active = $3,
			updated_at = $4
		WHERE id = $5
	`

	result, err := tx.Exec(
		ctx,
		query,
		trigger.NextRunAt,
		trigger.LastRunAt,
		trigger.IsActive,
		trigger.UpdatedAt,
		trigger.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return triggers.ErrTriggerNotFound
	}

	return nil
}

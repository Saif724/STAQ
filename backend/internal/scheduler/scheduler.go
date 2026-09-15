package scheduler

import (
	"context"
	"sync"

	"github.com/rs/zerolog"
)

type Scheduler struct {
	service *Service
	logger  zerolog.Logger

	mu      sync.Mutex
	running bool
}

func New(
	service *Service,
	logger zerolog.Logger,
) *Scheduler {
	return &Scheduler{
		service: service,
		logger:  logger,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()

	if s.running {
		s.mu.Unlock()

		s.logger.Warn().
			Msg("scheduler is already running")

		return
	}

	s.running = true

	s.mu.Unlock()

	go func() {
		if err := s.service.Run(ctx); err != nil {
			if err != context.Canceled && err != context.DeadlineExceeded {
				s.logger.Error().
					Err(err).
					Msg("scheduler exited with error")
			}
		}

		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()
}

// Package cron provides a scheduler for recurring jobs.
// It can either run jobs in-process or delegate to pg_cron when available.
package cron

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Job represents a scheduled job.
type Job struct {
	ID       string
	Schedule string // cron expression or "@every 1h"
	Task     func(ctx context.Context) error
}

// Scheduler manages scheduled jobs.
type Scheduler struct {
	mu       sync.RWMutex
	jobs     map[string]*Job
	timers   map[string]*time.Timer
	pgPool   *pgxpool.Pool
	usePgCron bool
}

// NewScheduler creates a new scheduler.
func NewScheduler(pgPool *pgxpool.Pool) *Scheduler {
	return &Scheduler{
		jobs:      make(map[string]*Job),
		timers:    make(map[string]*time.Timer),
		pgPool:    pgPool,
		usePgCron: pgPool != nil,
	}
}

// Register adds a job. If pg_cron is available, it attempts to install the job there;
// otherwise it falls back to an in-process ticker.
func (s *Scheduler) Register(ctx context.Context, job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[job.ID] = &job

	if s.usePgCron {
		// Try to schedule via pg_cron
		_, err := s.pgPool.Exec(ctx, "SELECT cron.schedule($1, $2, $3)", job.ID, job.Schedule, fmt.Sprintf("SELECT pg_notify('cron', '%s')", job.ID))
		if err != nil {
			log.Warn().Err(err).Str("job", job.ID).Msg("pg_cron unavailable, falling back to in-process scheduler")
			s.usePgCron = false
		} else {
			log.Info().Str("job", job.ID).Msg("Scheduled via pg_cron")
			return nil
		}
	}

	// In-process fallback: simple @every or rough cron support
	interval, err := parseSchedule(job.Schedule)
	if err != nil {
		return err
	}

	var run func()
	run = func() {
		if err := job.Task(ctx); err != nil {
			log.Error().Err(err).Str("job", job.ID).Msg("Cron job failed")
		}
		s.mu.Lock()
		if t, ok := s.timers[job.ID]; ok {
			t.Reset(interval)
		}
		s.mu.Unlock()
	}

	timer := time.AfterFunc(interval, run)
	s.timers[job.ID] = timer
	log.Info().Str("job", job.ID).Dur("interval", interval).Msg("Scheduled in-process")
	return nil
}

// Unregister removes a job.
func (s *Scheduler) Unregister(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.usePgCron {
		_, err := s.pgPool.Exec(ctx, "SELECT cron.unschedule($1)", id)
		if err != nil {
			log.Warn().Err(err).Str("job", id).Msg("Failed to unschedule pg_cron job")
		}
	}

	if t, ok := s.timers[id]; ok {
		t.Stop()
		delete(s.timers, id)
	}
	delete(s.jobs, id)
	return nil
}

// parseSchedule supports @every style and simple minutes.
func parseSchedule(schedule string) (time.Duration, error) {
	if schedule == "" {
		return 0, fmt.Errorf("empty schedule")
	}
	if strings.HasPrefix(schedule, "@every ") {
		d, err := time.ParseDuration(strings.TrimPrefix(schedule, "@every "))
		if err != nil {
			return 0, err
		}
		return d, nil
	}
	// Fallback: treat as minutes integer
	if mins, err := strconv.Atoi(schedule); err == nil {
		return time.Duration(mins) * time.Minute, nil
	}
	return 0, fmt.Errorf("unsupported schedule format: %s", schedule)
}

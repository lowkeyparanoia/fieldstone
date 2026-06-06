package jobs

import (
	"context"
	"sync"
	"time"
)

// MemoryStore is an in-memory implementation of JobStore for testing
type MemoryStore struct {
	mu    sync.RWMutex
	jobs  map[string]*Job
	queue map[string][]string // queue name -> job IDs
}

// NewMemoryStore creates a new in-memory job store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		jobs:  make(map[string]*Job),
		queue: make(map[string][]string),
	}
}

// CreateJob stores a new job
func (s *MemoryStore) CreateJob(ctx context.Context, job *Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[job.ID] = job
	s.queue[job.Queue] = append(s.queue[job.Queue], job.ID)
	return nil
}

// GetJob retrieves a job by ID
func (s *MemoryStore) GetJob(ctx context.Context, id string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if job, ok := s.jobs[id]; ok {
		j := *job
		return &j, nil
	}
	return nil, nil
}

// UpdateJob updates an existing job
func (s *MemoryStore) UpdateJob(ctx context.Context, job *Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[job.ID] = job
	return nil
}

// DeleteJob removes a job
func (s *MemoryStore) DeleteJob(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.jobs, id)
	return nil
}

// FetchAndLockJob retrieves and locks the next available job
func (s *MemoryStore) FetchAndLockJob(ctx context.Context, queue string, workerID string) (*Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	queueIDs, ok := s.queue[queue]
	if !ok {
		return nil, nil
	}

	for _, id := range queueIDs {
		job, ok := s.jobs[id]
		if !ok {
			continue
		}

		if job.Status == JobStatusPending && job.RunAt.Before(now) {
			job.Status = JobStatusRunning
			job.LockedAt = &now
			job.LockedBy = workerID
			job.Attempts++
			job.UpdatedAt = now
			return job, nil
		}
	}

	return nil, nil
}

// ListPendingJobs returns pending jobs
func (s *MemoryStore) ListPendingJobs(ctx context.Context, queue string, limit int) ([]*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var jobs []*Job
	queueIDs, ok := s.queue[queue]
	if !ok {
		return jobs, nil
	}

	for _, id := range queueIDs {
		if job, ok := s.jobs[id]; ok && job.Status == JobStatusPending {
			j := *job
			jobs = append(jobs, &j)
			if len(jobs) >= limit {
				break
			}
		}
	}

	return jobs, nil
}

// RetryFailedJobs retries failed jobs
func (s *MemoryStore) RetryFailedJobs(ctx context.Context, queue string, maxAttempts int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	retried := 0
	now := time.Now()

	for _, job := range s.jobs {
		if job.Queue == queue && job.Status == JobStatusFailed && job.Attempts < maxAttempts {
			job.Status = JobStatusPending
			job.RunAt = now.Add(time.Second * time.Duration(1<<job.Attempts))
			retried++
		}
	}

	return retried, nil
}

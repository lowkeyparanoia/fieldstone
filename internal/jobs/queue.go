// Package jobs provides a background job queue with goroutine workers.
// This implements the "Collection triggers/workflows" feature requested in PocketBase issue #317.
// It supports concurrent job processing with configurable worker pools.
package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// JobStatus represents the current state of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusRetrying  JobStatus = "retrying"
	JobStatusDead      JobStatus = "dead" // Dead letter queue
)

// Job represents a unit of work to be processed
type Job struct {
	ID          string          `json:"id"`
	Queue       string          `json:"queue"`
	Type        string          `json:"type"`        // e.g., "email", "webhook", "import"
	Payload     json.RawMessage `json:"payload"`
	Status      JobStatus       `json:"status"`
	Priority    int             `json:"priority"`    // Higher = more important
	Attempts    int             `json:"attempts"`
	MaxAttempts int             `json:"maxAttempts"`
	RunAt       time.Time       `json:"runAt"`       // Scheduled execution time
	LockedAt    *time.Time      `json:"lockedAt"`    // When a worker claimed it
	LockedBy    string          `json:"lockedBy"`    // Worker ID
	LastError   string          `json:"lastError"`
	Result      json.RawMessage `json:"result"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

// HandlerFunc processes a job
type HandlerFunc func(ctx context.Context, job *Job) error

// Queue manages job processing with worker pools
type Queue struct {
	mu          sync.RWMutex
	db          JobStore
	workers     map[string]*Worker
	handlers    map[string]HandlerFunc
	maxWorkers  int
	workerID    string
	shutdownCh  chan struct{}
	wg          sync.WaitGroup
}

// JobStore defines the interface for job persistence
type JobStore interface {
	CreateJob(ctx context.Context, job *Job) error
	GetJob(ctx context.Context, id string) (*Job, error)
	UpdateJob(ctx context.Context, job *Job) error
	DeleteJob(ctx context.Context, id string) error
	FetchAndLockJob(ctx context.Context, queue string, workerID string) (*Job, error)
	ListPendingJobs(ctx context.Context, queue string, limit int) ([]*Job, error)
	RetryFailedJobs(ctx context.Context, queue string, maxAttempts int) (int, error)
}

// Worker processes jobs from a queue
type Worker struct {
	id       string
	queue    string
	handler  HandlerFunc
	sem      chan struct{}     // Semaphore to limit concurrency
	stopCh   chan struct{}
	queueRef *Queue
}

// NewQueue creates a new job queue
func NewQueue(store JobStore, maxWorkers int) *Queue {
	if maxWorkers <= 0 {
		maxWorkers = 10
	}

	return &Queue{
		db:         store,
		workers:    make(map[string]*Worker),
		handlers:   make(map[string]HandlerFunc),
		maxWorkers: maxWorkers,
		workerID:   uuid.New().String(),
		shutdownCh: make(chan struct{}),
	}
}

// RegisterHandler registers a handler for a job type
func (q *Queue) RegisterHandler(jobType string, handler HandlerFunc) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[jobType] = handler
}

// Enqueue adds a job to the queue
func (q *Queue) Enqueue(ctx context.Context, jobType string, queue string, payload interface{}) (*Job, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	job := &Job{
		ID:          uuid.New().String(),
		Queue:       queue,
		Type:        jobType,
		Payload:     payloadJSON,
		Status:      JobStatusPending,
		Priority:    0,
		MaxAttempts: 3,
		RunAt:       time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := q.db.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	log.Info().
		Str("job_id", job.ID).
		Str("type", jobType).
		Str("queue", queue).
		Msg("Job enqueued")

	return job, nil
}

// EnqueueDelayed schedules a job for future execution
func (q *Queue) EnqueueDelayed(ctx context.Context, jobType string, queue string, payload interface{}, delay time.Duration) (*Job, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	job := &Job{
		ID:          uuid.New().String(),
		Queue:       queue,
		Type:        jobType,
		Payload:     payloadJSON,
		Status:      JobStatusPending,
		MaxAttempts: 3,
		RunAt:       time.Now().Add(delay),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := q.db.CreateJob(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}

// StartWorker starts a worker goroutine pool for a queue
func (q *Queue) StartWorker(queue string, concurrency int) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.workers[queue]; exists {
		return fmt.Errorf("worker for queue %s already running", queue)
	}

	if concurrency <= 0 {
		concurrency = 1
	}

	worker := &Worker{
		id:       fmt.Sprintf("%s-%s", q.workerID, queue),
		queue:    queue,
		sem:      make(chan struct{}, concurrency),
		stopCh:   make(chan struct{}),
		queueRef: q,
	}

	q.workers[queue] = worker

	// Start the worker goroutine
	q.wg.Add(1)
	go worker.run()

	log.Info().
		Str("queue", queue).
		Int("concurrency", concurrency).
		Msg("Worker started")

	return nil
}

// StopWorker stops a worker gracefully
func (q *Queue) StopWorker(queue string) error {
	q.mu.Lock()
	worker, exists := q.workers[queue]
	q.mu.Unlock()

	if !exists {
		return fmt.Errorf("no worker found for queue %s", queue)
	}

	close(worker.stopCh)
	delete(q.workers, queue)

	log.Info().Str("queue", queue).Msg("Worker stopped")
	return nil
}

// Shutdown gracefully stops all workers
func (q *Queue) Shutdown(ctx context.Context) error {
	log.Info().Msg("Shutting down job queue...")

	// Signal all workers to stop
	q.mu.Lock()
	for queue, worker := range q.workers {
		close(worker.stopCh)
		delete(q.workers, queue)
	}
	q.mu.Unlock()

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("All workers stopped gracefully")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout exceeded")
	}
}

// run is the main worker loop using goroutines
func (w *Worker) run() {
	defer w.queueRef.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.processJobs()
		}
	}
}

// processJobs fetches and processes jobs concurrently
func (w *Worker) processJobs() {
	ctx := context.Background()

	// Try to acquire a slot in the semaphore
	select {
	case w.sem <- struct{}{}:
		// Got a slot, process a job
	default:
		// Worker pool is full, skip this cycle
		return
	}

	// Fetch and lock a job
	job, err := w.queueRef.db.FetchAndLockJob(ctx, w.queue, w.id)
	if err != nil {
		<-w.sem // Release slot
		return
	}
	if job == nil {
		<-w.sem // Release slot
		return
	}

	// Process job in a goroutine
	go func() {
		defer func() { <-w.sem }() // Release slot when done

		w.executeJob(ctx, job)
	}()
}

// executeJob runs a single job with panic recovery
func (w *Worker) executeJob(ctx context.Context, job *Job) {
	// Get handler
	w.queueRef.mu.RLock()
	handler, exists := w.queueRef.handlers[job.Type]
	w.queueRef.mu.RUnlock()

	if !exists {
		job.Status = JobStatusFailed
		job.LastError = fmt.Sprintf("no handler registered for job type: %s", job.Type)
		job.UpdatedAt = time.Now()
		w.queueRef.db.UpdateJob(ctx, job)
		return
	}

	// Execute with panic recovery
	func() {
		defer func() {
			if r := recover(); r != nil {
				job.LastError = fmt.Sprintf("panic: %v", r)
				job.Status = JobStatusFailed
				job.Attempts++
				
				// Check if we should retry
				if job.Attempts < job.MaxAttempts {
					job.Status = JobStatusRetrying
					// Exponential backoff: 2^attempts seconds
					delay := time.Duration(1<<job.Attempts) * time.Second
					job.RunAt = time.Now().Add(delay)
					log.Info().
						Str("job_id", job.ID).
						Int("attempt", job.Attempts).
						Dur("retry_after", delay).
						Msg("Job failed, scheduling retry")
				} else {
					job.Status = JobStatusDead
					log.Error().
						Str("job_id", job.ID).
						Int("attempts", job.Attempts).
						Msg("Job moved to dead letter queue")
				}
				
				job.UpdatedAt = time.Now()
				w.queueRef.db.UpdateJob(ctx, job)
			}
		}()

		// Update job status
		job.Status = JobStatusRunning
		job.Attempts++
		job.UpdatedAt = time.Now()
		w.queueRef.db.UpdateJob(ctx, job)

		// Execute handler
		if err := handler(ctx, job); err != nil {
			job.LastError = err.Error()
			job.Status = JobStatusFailed
			
			if job.Attempts < job.MaxAttempts {
				job.Status = JobStatusRetrying
				delay := time.Duration(1<<job.Attempts) * time.Second
				job.RunAt = time.Now().Add(delay)
			} else {
				job.Status = JobStatusDead
			}
		} else {
			job.Status = JobStatusCompleted
			job.LastError = ""
			
			log.Info().
				Str("job_id", job.ID).
				Str("type", job.Type).
				Dur("duration", time.Since(job.LockedAt.Add(-time.Second))).
				Msg("Job completed successfully")
		}

		job.UpdatedAt = time.Now()
		w.queueRef.db.UpdateJob(ctx, job)
	}()
}

// GetQueueStats returns statistics for a queue
func (q *Queue) GetQueueStats(ctx context.Context, queue string) (*QueueStats, error) {
	// This would query the database for stats
	return &QueueStats{
		Queue:     queue,
		Pending:   0,
		Running:   0,
		Completed: 0,
		Failed:    0,
		Dead:      0,
	}, nil
}

// QueueStats contains queue statistics
type QueueStats struct {
	Queue     string `json:"queue"`
	Pending   int    `json:"pending"`
	Running   int    `json:"running"`
	Completed int    `json:"completed"`
	Failed    int    `json:"failed"`
	Dead      int    `json:"dead"`
}

// Common job types for PocketBase workflows
const (
	JobTypeEmailSend    = "email:send"
	JobTypeWebhookCall  = "webhook:call"
	JobTypeImportData   = "data:import"
	JobTypeExportData   = "data:export"
	JobTypeIndexRebuild = "index:rebuild"
	JobTypeNotifyUsers  = "users:notify"
)

package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/fieldstone/fieldstone/internal/api"
	"github.com/fieldstone/fieldstone/internal/auth"
	"github.com/fieldstone/fieldstone/internal/backend"
	"github.com/fieldstone/fieldstone/internal/cache"
	"github.com/fieldstone/fieldstone/internal/jobs"
)

func main() {
	// Parse flags
	var (
		port        = flag.String("port", "8090", "Server port")
		dsn         = flag.String("dsn", "fieldstone.db", "Database DSN")
		backendType = flag.String("backend", "sqlite", "Backend type: sqlite, postgres, memory")
		jwtSecret   = flag.String("jwt-secret", "", "JWT secret (default: auto-generated)")
	)
	flag.Parse()

	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Environment overrides must be applied BEFORE deciding whether a secret
	// needs generating, otherwise FIELDSTONE_JWT_SECRET is ignored on this path
	// and a fresh secret is minted on every boot, invalidating every token.
	if envPort := os.Getenv("FIELDSTONE_PORT"); envPort != "" {
		*port = envPort
	}
	if envDSN := os.Getenv("FIELDSTONE_DSN"); envDSN != "" {
		*dsn = envDSN
	}
	if envBackend := os.Getenv("FIELDSTONE_BACKEND"); envBackend != "" {
		*backendType = envBackend
	}
	if envJWT := os.Getenv("FIELDSTONE_JWT_SECRET"); envJWT != "" {
		*jwtSecret = envJWT
	}

	// Generate JWT secret if still not provided
	if *jwtSecret == "" {
		*jwtSecret = generateSecret()
		log.Warn().Msg("JWT secret auto-generated. Set FIELDSTONE_JWT_SECRET for persistence.")
	}

	log.Info().
		Str("backend", *backendType).
		Str("port", *port).
		Msg("Starting Fieldstone")

	// Initialize backend
	cfg := backend.Config{
		Type: *backendType,
		DSN:  *dsn,
	}

	if cfg.Type == "postgres" {
		cfg.MaxConns = 20
		cfg.MinConns = 5
	}

	be, err := backend.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize backend")
	}
	defer be.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := be.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping backend")
	}
	log.Info().Msg("Backend connected")

	// Initialize auth service
	authService := auth.NewService(*jwtSecret, 24*time.Hour, "fieldstone")

	// Initialize job queue with goroutine workers
	jobQueue := initJobQueue(be)
	defer jobQueue.Shutdown(context.Background())

	// Initialize cache
	cacheCfg := cache.Config{
		Type:       getEnv("CACHE_TYPE", "memory"),
		RedisAddr:  getEnv("CACHE_REDIS_ADDR", "localhost:6379"),
		RedisPass:  getEnv("CACHE_REDIS_PASSWORD", ""),
		RedisDB:    getEnvInt("CACHE_REDIS_DB", 0),
		DefaultTTL: getEnvDuration("CACHE_DEFAULT_TTL", 5*time.Minute),
		Prefix:     getEnv("CACHE_KEY_PREFIX", "fieldstone"),
	}

	var cacheInstance cache.Cache
	if cacheCfg.Type != "" && cacheCfg.Type != "none" {
		var err error
		cacheInstance, err = cache.New(cacheCfg)
		if err != nil {
			log.Error().Err(err).Msg("Failed to initialize cache, continuing without caching")
			cacheInstance = nil
		} else {
			log.Info().Str("type", cacheCfg.Type).Msg("Cache initialized")
			defer cacheInstance.Close()
		}
	}

	// Create API server
	server := api.NewServer(be, authService, jobQueue, cacheInstance)

	// Setup HTTP server
	srv := &http.Server{
		Addr:         ":" + *port,
		Handler:      server.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-quit:
			log.Info().Str("signal", sig.String()).Msg("Shutting down server...")

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			srv.SetKeepAlivesEnabled(false)
			if err := srv.Shutdown(ctx); err != nil {
				log.Fatal().Err(err).Msg("Server forced to shutdown")
			}

			close(done)
		}
	}()

	log.Info().Str("addr", srv.Addr).Msg("Server started")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Server failed to start")
	}

	select {
	case <-done:
		log.Info().Msg("Server stopped")
	}
}

// initJobQueue initializes the job queue with workers
func initJobQueue(be backend.Backend) *jobs.Queue {
	// Create an in-memory job store for MVP
	// In production, this would be a database-backed store
	store := jobs.NewMemoryStore()
	queue := jobs.NewQueue(store, 10)

	// Register job handlers
	queue.RegisterHandler(jobs.JobTypeEmailSend, func(ctx context.Context, job *jobs.Job) error {
		log.Info().Str("job_id", job.ID).Msg("Processing email job")
		// Email sending logic here
		return nil
	})

	queue.RegisterHandler(jobs.JobTypeWebhookCall, func(ctx context.Context, job *jobs.Job) error {
		log.Info().Str("job_id", job.ID).Msg("Processing webhook job")
		// Webhook call logic here
		return nil
	})

	queue.RegisterHandler(jobs.JobTypeImportData, func(ctx context.Context, job *jobs.Job) error {
		log.Info().Str("job_id", job.ID).Msg("Processing import job")
		// Data import logic here
		return nil
	})

	// Start workers with goroutines
	if err := queue.StartWorker("default", 5); err != nil {
		log.Error().Err(err).Msg("Failed to start default worker")
	}
	if err := queue.StartWorker("emails", 3); err != nil {
		log.Error().Err(err).Msg("Failed to start email worker")
	}
	if err := queue.StartWorker("webhooks", 10); err != nil {
		log.Error().Err(err).Msg("Failed to start webhook worker")
	}

	return queue
}

// generateSecret creates a random secret for JWT signing
// generateSecret returns 32 bytes of cryptographically secure randomness.
//
// The previous implementation called os.ReadFile("/dev/urandom"). ReadFile
// reads a file to EOF, and /dev/urandom is an endless stream, so it never
// returned: the process spun forever allocating unbounded memory. Because the
// environment overrides were applied after this call, *jwtSecret was always
// still empty here, so the server hung on every start unless -jwt-secret was
// passed as a command line flag.
//
// crypto/rand is the correct source. It is the OS CSPRNG, it never blocks on
// modern kernels, and it reports failure instead of silently returning zeros,
// which the old code did whenever the ReadFile probe failed.
func generateSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatal().Err(err).Msg("Failed to read from the system CSPRNG")
	}
	return fmt.Sprintf("%x", b)
}

// getEnv gets environment variable or returns default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets environment variable as int or returns default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvDuration gets environment variable as duration or returns default
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

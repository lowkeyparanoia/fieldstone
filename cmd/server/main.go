package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/fieldstone/fieldstone/internal/api"
	"github.com/fieldstone/fieldstone/internal/auth"
	"github.com/fieldstone/fieldstone/internal/backend"
	"github.com/fieldstone/fieldstone/internal/cache"
	"github.com/fieldstone/fieldstone/internal/cron"
	"github.com/fieldstone/fieldstone/internal/functions"
	"github.com/fieldstone/fieldstone/internal/jobs"
	wasm "github.com/fieldstone/fieldstone/internal/plugins"
	"github.com/fieldstone/fieldstone/internal/realtime"
	"github.com/fieldstone/fieldstone/internal/storage"
	"github.com/fieldstone/fieldstone/internal/webhooks"
)

func main() {
	// Parse flags
	var (
		port           = flag.String("port", "8090", "Server port")
		dsn            = flag.String("dsn", "fieldstone.db", "Database DSN")
		backendType    = flag.String("backend", "sqlite", "Backend type: sqlite, postgres, memory")
		jwtSecret      = flag.String("jwt-secret", "", "JWT secret (default: auto-generated)")
		storageType    = flag.String("storage-type", getEnv("STORAGE_TYPE", "local"), "Storage type: local, s3, none")
		functionsHost  = flag.String("functions-host", getEnv("FUNCTIONS_HOST", ""), "Functions sidecar host")
	)
	flag.Parse()

	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Generate JWT secret if not provided
	if *jwtSecret == "" {
		*jwtSecret = generateSecret()
		log.Warn().Msg("JWT secret auto-generated. Set FIELDSTONE_JWT_SECRET for persistence.")
	}

	// Override with env vars
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

	// Extract pgx pool if using postgres (for gateway, RLS, cron)
	var pgPool *pgxpool.Pool
	if pgBe, ok := be.(*backend.PostgresBackend); ok {
		pgPool = pgBe.Pool()
	}

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

	// Initialize storage
	var storageBackend storage.Backend
	if *storageType != "none" {
		storageCfg := storage.Config{
			Type:      *storageType,
			LocalPath: getEnv("STORAGE_LOCAL_PATH", "./storage"),
		}
		storageBackend, err = storage.New(storageCfg)
		if err != nil {
			log.Error().Err(err).Msg("Failed to initialize storage, continuing without storage")
		} else {
			log.Info().Str("type", *storageType).Msg("Storage initialized")
		}
	}

	// Initialize realtime manager
	rtManager := realtime.NewManager()
	go rtManager.Run()
	log.Info().Msg("Realtime manager started")

	// Initialize webhooks
	whStore := webhooks.NewMemoryStore()
	whDeliveryStore := webhooks.NewMemoryDeliveryStore()
	whManager := webhooks.NewManager(whStore, whDeliveryStore)
	whManager.Start(3)
	log.Info().Msg("Webhook manager started")

	// Initialize functions proxy
	var fnProxy *functions.Proxy
	if *functionsHost != "" {
		fnProxy, err = functions.NewProxy(*functionsHost)
		if err != nil {
			log.Error().Err(err).Msg("Failed to initialize functions proxy")
		} else {
			log.Info().Str("host", *functionsHost).Msg("Functions proxy initialized")
		}
	}

	// Initialize cron scheduler
	scheduler := cron.NewScheduler(pgPool)
	// Register some example jobs
	_ = scheduler.Register(ctx, cron.Job{
		ID:       "heartbeat",
		Schedule: "@every 5m",
		Task: func(ctx context.Context) error {
			log.Debug().Msg("Cron heartbeat")
			return nil
		},
	})

	// Create API server
	server := api.NewServer(be, authService, jobQueue, cacheInstance, pgPool, storageBackend, fnProxy)
	server.SetRealtimeManager(rtManager)
	server.SetWebhookManager(whManager)

	// WASM functions runtime — in-process, sandboxed transforms via wazero.
	// Loads any *.wasm under the plugins dir (set FIELDSTONE_WASM_DIR; defaults
	// to the bundled slugify example) and exposes them at /api/functions/wasm/*.
	if pluginMgr, err := wasm.NewManager(context.Background()); err != nil {
		log.Error().Err(err).Msg("Failed to init WASM plugin manager")
	} else {
		wasmDir := getEnv("FIELDSTONE_WASM_DIR", "internal/plugins/examples/slugify")
		loaded := 0
		if entries, derr := os.ReadDir(wasmDir); derr == nil {
			for _, e := range entries {
				if e.IsDir() || filepath.Ext(e.Name()) != ".wasm" {
					continue
				}
				bytes, rerr := os.ReadFile(filepath.Join(wasmDir, e.Name()))
				if rerr != nil {
					continue
				}
				id := strings.TrimSuffix(e.Name(), ".wasm")
				if _, lerr := pluginMgr.Load(id, id, "1.0.0", bytes); lerr != nil {
					log.Warn().Err(lerr).Str("plugin", id).Msg("Failed to load WASM plugin")
					continue
				}
				loaded++
			}
		}
		log.Info().Int("loaded", loaded).Str("dir", wasmDir).Msg("WASM functions runtime ready")
		server.SetPluginManager(pluginMgr)
		defer pluginMgr.Close()
	}

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
func generateSecret() string {
	// crypto/rand reads exactly 32 bytes from the OS CSPRNG. (The previous
	// implementation did os.ReadFile("/dev/urandom"), which tries to read the
	// entire — infinite — device and hangs the server on startup.)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatal().Err(err).Msg("Failed to generate JWT secret")
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

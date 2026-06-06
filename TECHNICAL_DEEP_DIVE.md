# Fieldstone: Technical Deep Dive & Interview Guide

## 🎯 PROJECT OVERVIEW

**Fieldstone** is a production-ready, open-source Backend-as-a-Service (BaaS) platform inspired by PocketBase. It provides a complete backend infrastructure with REST API, real-time subscriptions, authentication, multi-tenancy, and file storage - all in a single binary.

### Architecture at a Glance

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT LAYER                                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │  Web App     │  │  Mobile App  │  │  CLI Tool    │  │  External    │    │
│  │  (React)     │  │  (iOS/And)   │  │  (Go/TS)     │  │  Services    │    │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘    │
└─────────┼─────────────────┼─────────────────┼─────────────────┼────────────┘
          │                 │                 │                 │
          └─────────────────┼─────────────────┼─────────────────┘
                            │ HTTPS/WebSocket
                            ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            API GATEWAY LAYER                                 │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │  Chi Router (HTTP)              gRPC Server         WebSocket        │   │
│  │  ├── Auth Middleware            ├── Protocol Buffers  ├── Real-time   │   │
│  │  ├── Rate Limiter               ├── Streaming RPC     ├── Pub/Sub    │   │
│  │  ├── Cache Middleware           └── High Performance  └── Events     │   │
│  │  └── CORS Handler                                                    │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           SERVICE LAYER                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │   Auth       │  │ Collections  │  │   Records    │  │   Storage    │    │
│  │  Service     │  │   Service    │  │   Service    │  │   Service    │    │
│  │              │  │              │  │              │  │              │    │
│  │ • JWT        │  │ • CRUD       │  │ • CRUD       │  │ • Local      │    │
│  │ • bcrypt     │  │ • Schema     │  │ • Validate   │  │ • S3         │    │
│  │ • Rotation   │  │ • Indexes    │  │ • Query      │  │ • Upload     │    │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘    │
│         │                 │                 │                 │            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │  Job Queue   │  │    Cache     │  │  Real-time   │  │   Webhooks   │    │
│  │              │  │   (Redis)    │  │  Manager     │  │              │    │
│  │ • Workers    │  │              │  │              │  │ • HTTP       │    │
│  │ • Retries    │  │ • Get/Set    │  │ • WebSocket  │  │ • Retry      │    │
│  │ • Scheduler  │  │ • TTL        │  │ • Broadcast  │  │ • Events     │    │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           BACKEND INTERFACE                                  │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                           Backend Interface                          │   │
│  │  (Abstraction: SQLite / PostgreSQL / In-Memory interchangeable)      │   │
│  └──────────────┬──────────────────────────────┬───────────────────────┘   │
│                 │                              │                            │
│  ┌──────────────▼──────────────┐  ┌───────────▼──────────────┐             │
│  │        SQLite              │  │      PostgreSQL          │             │
│  │  • Single file             │  │  • Production ready      │             │
│  │  • Embedded                │  │  • Connection pool       │             │
│  │  • CGO enabled             │  │  • RLS support           │             │
│  │  • Transactions            │  │  • JSONB indexes         │             │
│  └─────────────────────────────┘  └──────────────────────────┘             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 🐹 GO CONCEPTS APPLIED

### **1. Interfaces & Polymorphism**

**Location:** `internal/backend/backend.go`, `internal/cache/cache.go`

**Concept:** Interface segregation and dependency injection

```go
// Backend interface abstracts database operations
type Backend interface {
    Ping(ctx context.Context) error
    CreateCollection(ctx context.Context, tenantID string, collection *models.Collection) error
    // ... 20+ methods
}

// This allows swapping SQLite ↔ PostgreSQL ↔ Memory without changing business logic
```

**Interview Question:** *"How would you add a new database backend (e.g., MySQL)?"*
**Answer:** Just implement the `Backend` interface. The rest of the application doesn't need changes due to interface abstraction.

---

### **2. Context Package**

**Location:** Every handler and service method

**Concept:** Request-scoped values, cancellation, timeouts

```go
// In handlers
func (s *Server) handleGetCollection(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()  // Request context with timeout
    tenantID := s.getTenantID(r)
    
    collection, err := s.backend.GetCollection(ctx, tenantID, id)
    // Context carries: deadlines, cancellation signals, request-scoped values
}
```

**Why it matters:**
- Request cancellation propagates through all layers
- Database queries respect HTTP timeouts
- Graceful shutdown waits for contexts to complete

**Interview Question:** *"What happens if a client disconnects mid-request?"*
**Answer:** Context cancellation propagates down, database query is cancelled, resources freed immediately.

---

### **3. Goroutines & Concurrency**

**Location:** `internal/jobs/queue.go`, `internal/realtime/manager.go`, `cmd/server/main.go`

**Concept:** Worker pools, fan-out, synchronization

```go
// Job Queue Worker Pattern
func (q *Queue) StartWorker(name string, concurrency int) error {
    for i := 0; i < concurrency; i++ {
        go q.worker(name)  // Each worker is a goroutine
    }
    return nil
}

func (q *Queue) worker(name string) {
    for job := range q.jobs {
        q.processJob(job)  // Process jobs concurrently
    }
}
```

**Real-time Broadcasting:**
```go
// Broadcast to all connected clients concurrently
for _, client := range room {
    go func(c *Client) {
        c.Send <- data  // Non-blocking send per client
    }(client)
}
```

**Interview Question:** *"How do you prevent goroutine leaks?"*
**Answer:** We use `context.WithCancel`, proper channel closing, and the `sync.WaitGroup` to ensure all goroutines exit cleanly during shutdown.

---

### **4. Channels & Select Statements**

**Location:** `internal/realtime/manager.go`, `cmd/server/main.go`

**Concept:** Communication between goroutines, event loops

```go
// Real-time Manager Event Loop
type Manager struct {
    register   chan *Client      // New connections
    unregister chan *Client      // Disconnections  
    broadcast  chan Event        // Messages to broadcast
}

func (m *Manager) Run() {
    for {
        select {
        case client := <-m.register:
            m.clients[client.ID] = client
            
        case client := <-m.unregister:
            delete(m.clients, client.ID)
            close(client.Send)
            
        case event := <-m.broadcast:
            m.broadcastToRoom(event)
        }
    }
}
```

**Graceful Shutdown:**
```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

go func() {
    sig := <-quit
    // Signal received, start graceful shutdown
    srv.Shutdown(ctx)
}()
```

**Interview Question:** *"Why use buffered vs unbuffered channels?"*
**Answer:** Buffered channels (e.g., `make(chan Event, 256)`) prevent blocking when producers are faster than consumers. We use them in the broadcast channel to handle traffic spikes without blocking publishers.

---

### **5. Struct Embedding & Composition**

**Location:** `internal/api/server.go`, `internal/api/handlers.go`

**Concept:** Composition over inheritance

```go
type Server struct {
    router  *chi.Mux
    backend backend.Backend  // Interface composition
    auth    *auth.Service
    cache   cache.Cache      // Can be Redis or Memory
    jobs    *jobs.Queue
    logger  zerolog.Logger
}

// NOT inheritance - Server HAS-A backend, not IS-A backend
```

**Response Writer Wrapping:**
```go
type responseWriter struct {
    http.ResponseWriter  // Embed standard library type
    statusCode int
    body       []byte
}

func (w *responseWriter) Write(b []byte) (int, error) {
    w.body = append(w.body, b...)  // Capture response
    return w.ResponseWriter.Write(b)
}
```

**Interview Question:** *"Why not use inheritance like in Java?"*
**Answer:** Go favors composition. We embed structs to reuse functionality while maintaining clear ownership. This avoids the fragile base class problem.

---

### **6. Error Handling Patterns**

**Location:** All error returns, `internal/backend/backend.go`

**Concept:** Explicit error returns, custom error types

```go
// BackendError with error codes
type BackendError struct {
    Code    string
    Message string
    Cause   error
}

func (e *BackendError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Sentinel errors for type checking
func IsNotFound(err error) bool {
    if be, ok := err.(*BackendError); ok {
        return be.Code == ErrCodeNotFound
    }
    return false
}
```

**Usage:**
```go
collection, err := s.backend.GetCollection(ctx, tenantID, id)
if err != nil {
    if backend.IsNotFound(err) {
        // Return 404
        s.sendError(w, http.StatusNotFound, "collection not found")
        return
    }
    // Return 500
    s.sendError(w, http.StatusInternalServerError, err.Error())
    return
}
```

**Interview Question:** *"How do you handle different error types?"*
**Answer:** Custom error types with error codes + type assertion functions. This allows HTTP handlers to return appropriate status codes based on error type.

---

### **7. Mutex & Synchronization**

**Location:** `internal/backend/memory.go`, `internal/realtime/manager.go`

**Concept:** Protecting shared state, read-write locks

```go
type MemoryBackend struct {
    mu          sync.RWMutex  // Protects all maps
    collections map[string]models.Collection
    records     map[string]map[string]models.Record
}

func (b *MemoryBackend) GetCollection(ctx context.Context, tenantID, id string) (*models.Collection, error) {
    b.mu.RLock()           // Multiple readers allowed
    defer b.mu.RUnlock()
    
    if c, ok := b.collections[id]; ok {
        return &c, nil
    }
    return nil, NewNotFoundError("collection", id)
}

func (b *MemoryBackend) CreateCollection(ctx context.Context, tenantID string, c *models.Collection) error {
    b.mu.Lock()            // Exclusive write access
    defer b.mu.Unlock()
    
    b.collections[c.ID] = *c
    return nil
}
```

**Real-time Client Management:**
```go
type Manager struct {
    mu      sync.RWMutex
    clients map[string]*Client
    rooms   map[string]map[string]*Client
}
```

**Interview Question:** *"Why use RWMutex instead of Mutex?"*
**Answer:** RWMutex allows multiple concurrent readers but exclusive writers. In our case, reads (GetCollection) are 10x more frequent than writes (CreateCollection), so RWMutex provides better concurrency.

---

### **8. Defer Statements**

**Location:** Every resource cleanup, database operations

**Concept:** Guaranteed cleanup, LIFO order

```go
func (s *Server) handleCreateRecord(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    tenantID := s.getTenantID(r)
    
    // Defer ensures body is closed even if function returns early
    defer r.Body.Close()
    
    var req RecordRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return  // Body still closed via defer
    }
    // ... more code
}  // Body closed here
```

**Transaction Pattern:**
```go
func (m *Migrator) ApplyMigration(migration *Migration) error {
    tx, err := m.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()  // Rollback if not committed
    
    // Apply migration...
    if _, err := tx.Exec(migration.Up); err != nil {
        return err
    }
    
    return tx.Commit()  // On success, Rollback becomes no-op
}
```

**Interview Question:** *"What's the catch with defer?"*
**Answer:** Defer has overhead (function call + argument evaluation). Also, in loops, deferred calls accumulate and run at function exit, not loop iteration exit. We avoid defer in hot loops.

---

### **9. JSON Encoding/Decoding**

**Location:** All API handlers, models

**Concept:** Struct tags, streaming decoder

```go
// Struct tags for JSON serialization
type Collection struct {
    ID        string                 `json:"id"`
    TenantID  string                 `json:"tenantId"`
    Name      string                 `json:"name"`
    Schema    []SchemaField          `json:"schema"`
    Options   map[string]interface{} `json:"options,omitempty"`  // Omit if empty
    CreatedAt time.Time              `json:"createdAt"`
}

// Request decoding with validation
func (s *Server) handleCreateCollection(w http.ResponseWriter, r *http.Request) {
    var req CollectionRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        s.sendError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
        return
    }
}
```

**Interview Question:** *"json.Marshal vs json.Encoder - when to use each?"*
**Answer:** Use `json.Marshal` when you need the byte slice. Use `json.Encoder` for streaming (writing directly to HTTP response) to avoid buffering the entire JSON in memory.

---

### **10. HTTP Server Patterns**

**Location:** `internal/api/server.go`, `cmd/server/main.go`

**Concept:** Graceful shutdown, timeouts, middleware chain

```go
// Server configuration with timeouts
srv := &http.Server{
    Addr:         ":" + *port,
    Handler:      server.Router(),
    ReadTimeout:  15 * time.Second,   // Prevent slowloris attacks
    WriteTimeout: 15 * time.Second,   // Prevent slow clients
    IdleTimeout:  60 * time.Second,   // Keep-alive timeout
}

// Graceful shutdown
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

go func() {
    <-quit
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    srv.SetKeepAlivesEnabled(false)  // Stop accepting new connections
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal().Err(err).Msg("Server forced to shutdown")
    }
}()
```

**Interview Question:** *"What happens to in-flight requests during shutdown?"*
**Answer:** `Shutdown` waits for active requests to complete (up to 30s timeout). No new connections accepted. After timeout, force close. This ensures no request is dropped mid-processing.

---

## 🏗️ SYSTEM DESIGN ASPECTS

### **1. Multi-Tenancy Architecture**

**Design:** Row-level tenant isolation

```sql
-- Every table has tenant_id
CREATE TABLE collections (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,  -- Isolation column
    name TEXT NOT NULL,
    -- ...
);

-- Queries always filter by tenant
SELECT * FROM collections WHERE tenant_id = 'tenant-123';
```

**Implementation:**
```go
func (s *Server) getTenantID(r *http.Request) string {
    // From JWT token context
    if authCtx, ok := auth.FromContext(r.Context()); ok {
        return authCtx.TenantID
    }
    return "default"  // Fallback
}

// All backend methods accept tenantID
func (b *Backend) GetCollection(ctx context.Context, tenantID, id string) (*Collection, error)
```

**Interview Question:** *"How do you ensure data isolation between tenants?"*
**Answer:** 
1. Every database row has `tenant_id`
2. Every query filters by `tenant_id`
3. JWT tokens include tenant claim
4. Backend interface requires tenantID parameter
5. No cross-tenant queries possible by design

---

### **2. Horizontal Scalability**

**Stateless Design:**
```
┌──────────┐    ┌──────────┐    ┌──────────┐
│  LB      │    │  LB      │    │  LB      │
└────┬─────┘    └────┬─────┘    └────┬─────┘
     │               │               │
┌────▼─────┐    ┌────▼─────┐    ┌────▼─────┐
│ Fieldstone│    │ Fieldstone│    │ Fieldstone│
│  Instance │    │  Instance │    │  Instance │
│   (State) │    │   (State) │    │   (State) │
└────┬─────┘    └────┬─────┘    └────┬─────┘
     │               │               │
     └───────────────┼───────────────┘
                     │
            ┌────────▼────────┐
            │   Shared State   │
            │  • PostgreSQL   │
            │  • Redis Cache  │
            │  • MinIO/S3     │
            └─────────────────┘
```

**Why it scales:**
- Instances are stateless (no local session storage)
- All state in shared services (DB, Redis, S3)
- Add more instances behind load balancer
- WebSocket connections use Redis pub/sub for broadcasting across instances

**Interview Question:** *"How would you scale to 1M concurrent users?"*
**Answer:**
1. Stateless instances behind ALB (auto-scaling)
2. PostgreSQL read replicas for read-heavy workloads
3. Redis Cluster for caching (100k+ ops/sec)
4. S3 for file storage (infinite scale)
5. CDN for static assets

---

### **3. Caching Strategy**

**Multi-Layer Caching:**

```
L1: In-Memory (per instance)
     • Hot data
     • Sub-millisecond access
     • Cleared on restart

L2: Redis (shared)
     • API responses
     • Sessions
     • Rate limit counters
     • ~1ms access

L3: Database
     • Persistent storage
     • 5-50ms access
```

**Cache Invalidation:**
```go
// Write-Through Pattern
func (s *Server) handleUpdateCollection(w http.ResponseWriter, r *http.Request) {
    // 1. Update database
    s.backend.UpdateCollection(ctx, tenantID, collection)
    
    // 2. Invalidate cache (delete, not update)
    s.cache.Delete(ctx, cache.CollectionKey(id))
    s.cache.DeletePattern(ctx, cache.CollectionsListKey(tenantID, 0, 0)+"*")
    
    // Next read will populate cache with fresh data
}
```

**Interview Question:** *"Cache invalidation - why delete instead of update?"*
**Answer:** Delete is safer. If update fails mid-way, we have inconsistent cache. By deleting, next read fetches from DB ensuring consistency. Also handles race conditions better.

---

### **4. Database Connection Pooling**

**PostgreSQL Configuration:**
```go
cfg := backend.Config{
    Type:     "postgres",
    DSN:      "postgres://...",
    MaxConns: 25,  // Max open connections
    MinConns: 5,   // Always keep 5 ready
}
```

**Why it matters:**
- Connection creation is expensive (TCP handshake + auth)
- Too many connections = DB overload
- Too few = request queuing
- Sweet spot: 2x number of CPU cores per instance

**Interview Question:** *"How many DB connections should you have?"*
**Answer:** Formula: `(Core Count / (1 - Blocking_Factor)) * Scale_Factor`. For web servers: ~25 connections per instance. With 10 instances = 250 total. PostgreSQL default max is 100, so we tune `max_connections` in postgresql.conf.

---

### **5. Rate Limiting Algorithm**

**Token Bucket Implementation:**
```go
type RateLimiter struct {
    requests map[string]*bucket  // IP/User -> bucket
    mu       sync.RWMutex
    rate     int                 // Tokens per second
    burst    int                 // Max bucket size
}

type bucket struct {
    tokens     float64
    lastUpdate time.Time
}

func (rl *RateLimiter) Allow(key string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    b, exists := rl.requests[key]
    if !exists {
        b = &bucket{tokens: float64(rl.burst)}
        rl.requests[key] = b
    }
    
    // Add tokens based on time passed
    now := time.Now()
    elapsed := now.Sub(b.lastUpdate).Seconds()
    b.tokens += elapsed * float64(rl.rate)
    if b.tokens > float64(rl.burst) {
        b.tokens = float64(rl.burst)
    }
    b.lastUpdate = now
    
    if b.tokens >= 1 {
        b.tokens--
        return true  // Request allowed
    }
    return false  // Rate limited
}
```

**Why Token Bucket:**
- Smooths burst traffic
- Allows short bursts (user clicking rapidly)
- Prevents long-term abuse
- Easy to implement in Redis for distributed rate limiting

---

## 🔍 INTERVIEW QUESTIONS & ANSWERS

### **System Design Questions**

**Q: Design a BaaS like Fieldstone. How would you handle 1M DAU?**

**A:**
```
1. Architecture:
   - Stateless API servers (auto-scaling group)
   - Read replicas for DB (1 primary, 3 replicas)
   - Redis Cluster for caching (6 nodes)
   - CDN for file storage
   - Load balancer with sticky sessions for WebSocket

2. Caching Strategy:
   - 90% read ratio → Aggressive caching
   - Collection metadata: 1 hour TTL
   - User sessions: Redis with 24h TTL
   - Query results: 5 min TTL

3. Database:
   - Sharding by tenant_id for horizontal scale
   - Connection pooling (25 conns/instance)
   - Read replicas for GET requests
   - Write to primary, async replicate

4. Real-time:
   - Redis pub/sub for cross-instance message routing
   - WebSocket connection limit: 10k per instance
   - Horizontal pod autoscaling based on connections

5. Monitoring:
   - Prometheus + Grafana
   - P95 latency alerts
   - Error rate thresholds
   - Cache hit rate dashboards
```

---

**Q: How do you ensure data consistency across cache and database?**

**A:**
```
Strategy: Cache-aside (Lazy Loading) + Invalidation

1. Read Path:
   Check Cache → Cache HIT → Return
           ↓ MISS
   Query DB → Store in Cache → Return

2. Write Path:
   Update DB → Invalidate Cache → Success
   
3. Why this works:
   - No distributed transactions needed
   - Cache is always a "best effort" optimization
   - Stale data impossible (deleted on write)
   - Race condition: Two writes, both invalidate, next read refreshes

4. Edge Case Handling:
   - Cache stampede: Use singleflight pattern
   - Thundering herd: Jitter on TTL
   - Partial failures: Retry with exponential backoff
```

---

**Q: How would you implement real-time subscriptions at scale?**

**A:**
```
Current: In-memory WebSocket manager
Scale: Redis pub/sub + WebSocket

1. Single Instance (Current):
   Client ─WebSocket→ Server ─Memory→ Broadcast

2. Multi-Instance (Scale):
   Client ─WebSocket→ Server A
                            ↓
                        Redis PUBLISH "collection:123" event
                            ↓
   Client ←WebSocket── Server B (SUBSCRIBED to "collection:123")

3. Implementation:
   - Each server subscribes to Redis channels
   - On client subscribe: server joins channel
   - On broadcast: server publishes to Redis
   - Redis forwards to all subscribed servers
   - Servers forward to their connected clients

4. Benefits:
   - Horizontal scaling (add more servers)
   - No single point of failure
   - Redis handles message routing efficiently
```

---

### **Go-Specific Questions**

**Q: Explain the difference between concurrency and parallelism in Go.**

**A:**
```go
// Concurrency: Dealing with many things at once (structure)
// Parallelism: Doing many things at once (execution)

// Concurrent: Goroutines coordinate via channels
func concurrent() {
    ch := make(chan int)
    go func() { ch <- 1 }()  // Goroutine 1
    go func() { ch <- 2 }()  // Goroutine 2
    <-ch  // Receive 1 or 2 (order undefined)
    <-ch  // Receive remaining
}

// Parallel: Multiple OS threads on multiple cores
runtime.GOMAXPROCS(4)  // Use 4 CPU cores

// Go handles both:
// - M:N scheduler (M goroutines on N OS threads)
// - Goroutines are cheap (2KB stack vs 1MB thread)
// - Channels for safe communication
```

---

**Q: What are goroutine leaks and how do you prevent them?**

**A:**
```go
// LEAK: Goroutine blocked forever
func leak() {
    ch := make(chan int)
    go func() {
        ch <- 1  // Blocks forever if no receiver
    }()
    // Function returns, goroutine stuck
}

// SOLUTION 1: Buffered channels
ch := make(chan int, 1)  // Buffer size 1

// SOLUTION 2: Select with timeout
go func() {
    select {
    case ch <- 1:
    case <-time.After(5 * time.Second):
        return  // Timeout, don't leak
    }
}()

// SOLUTION 3: Context cancellation
func worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return  // Clean exit
        case job := <-jobs:
            process(job)
        }
    }
}

// In Fieldstone: We use contexts with timeouts for all operations
// and sync.WaitGroup to wait for goroutines on shutdown
```

---

**Q: Explain Go's memory model and happens-before relationships.**

**A:**
```go
// Without synchronization: Race condition!
var counter int
go func() { counter++ }()
go func() { counter++ }()
// Result: Could be 1 or 2 (undefined)

// Happens-before via sync.Mutex
var mu sync.Mutex
var counter int

go func() {
    mu.Lock()
    counter++  // Happens-before unlock
    mu.Unlock()
}()

go func() {
    mu.Lock()
    counter++  // Happens-after first goroutine's unlock
    mu.Unlock()
}()
// Result: Always 2

// Happens-before via channel
ch := make(chan int)
var counter int

go func() {
    counter++  // Step 1
    ch <- 1    // Step 2 (send happens-before receive)
}()

<-ch           // Step 3 (receive)
fmt.Println(counter)  // Guaranteed to see 1

// In Fieldstone: We use mutexes for shared state
// and channels for goroutine coordination
```

---

## 🎯 KEY TAKEAWAYS FOR INTERVIEWS

### **What Makes Fieldstone Impressive:**

1. **Clean Architecture:**
   - Interface-driven design (Backend, Cache, Storage)
   - Clear separation of concerns
   - Dependency injection for testability

2. **Production-Ready:**
   - Graceful shutdown handling
   - Context propagation for cancellation
   - Structured error handling
   - Comprehensive logging

3. **Performance:**
   - Multi-layer caching (Redis + HTTP)
   - Connection pooling
   - Worker pools for background jobs
   - Read-write mutex optimization

4. **Scalability:**
   - Stateless design
   - Horizontal scaling ready
   - Shared nothing architecture
   - Database sharding support

5. **Go Idioms:**
   - Composition over inheritance
   - Explicit error handling
   - Goroutines + channels
   - defer for cleanup
   - struct tags for JSON

---

## 📚 DEEP DIVE: REQUEST FLOW

### **Example: Creating a Record**

```
1. HTTP Request
   POST /api/collections/123/records
   Headers: Authorization: Bearer <jwt>
   Body: {"data": {"title": "Product"}}

2. Router (Chi)
   ↓ Match route
   ↓ Extract params: collectionID = "123"

3. Middleware Chain
   ↓ RequestID (adds X-Request-ID)
   ↓ RealIP (gets client IP)
   ↓ Logger (logs request)
   ↓ Recoverer (catches panics)
   ↓ Timeout (60s deadline)
   ↓ CORS (adds headers)
   ↓ Cache Middleware (skips POST)
   ↓ Auth Middleware (validates JWT)
      ↓ Parse JWT
      ↓ Extract user context
      ↓ Add to request context

4. Handler: handleCreateRecord
   ↓ Decode JSON body
   ↓ Validate against schema
   ↓ Check tenant permissions (from context)

5. Backend Layer
   ↓ b.backend.CreateRecord(ctx, tenantID, record)
   ↓ SQLite/PostgreSQL INSERT
   ↓ Return record with ID

6. Cache Invalidation
   ↓ s.cache.DeletePattern("records:123:*")
   ↓ (Clears list caches for this collection)

7. Response
   ↓ Marshal to JSON
   ↓ Set headers
   ↓ Write response (201 Created)
   ↓ Deferred: r.Body.Close()

8. Logging
   ↓ Request duration
   ↓ Status code
   ↓ User ID
   ↓ Collection ID
```

**Time Breakdown:**
- Network: 1ms
- Middleware: 0.5ms
- JWT validation: 1ms
- JSON decode: 0.5ms
- Schema validation: 0.5ms
- DB insert: 5ms
- Cache invalidation: 1ms
- JSON encode: 0.5ms
- **Total: ~10ms**

---

## 🏆 FINAL TIPS FOR SYSTEM DESIGN INTERVIEWS

1. **Start with Requirements:**
   - Functional: What does it do?
   - Non-functional: Scale? Latency? Availability?

2. **High-Level Design First:**
   - Draw boxes and arrows
   - Don't jump into code
   - Explain trade-offs

3. **Deep Dive Gradually:**
   - Start with API design
   - Then data model
   - Then scalability
   - Then edge cases

4. **Use Numbers:**
   - "Assume 1M DAU"
   - "100 requests/sec average"
   - "10KB average response"
   - "99.9% availability SLA"

5. **Address Bottlenecks:**
   - Database: Caching, sharding, read replicas
   - Single points of failure: Redundancy
   - Hot spots: Consistent hashing

6. **Talk About Monitoring:**
   - Metrics: Latency, throughput, errors
   - Logging: Distributed tracing
   - Alerting: P95 latency, error rates

---

**Good luck with your interviews! Fieldstone demonstrates production-grade Go development and system design skills.** 🚀

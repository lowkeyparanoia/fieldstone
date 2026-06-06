# Fieldstone Testing Guide for Portfolio

## Quick Test Commands for Screenshots

### 1. Build & Run (Terminal Screenshot)
```bash
cd /Users/Jre/fieldstone
export PATH="/Users/Jre/go-install/go/bin:$PATH"
go version
go mod tidy
go build -o bin/fieldstone ./cmd/server
./bin/fieldstone
```

**Screenshot 1**: Build success message showing binary creation
**Screenshot 2**: Server startup showing "Starting Fieldstone" and port 8090

### 2. Health Check (Browser/Terminal)
```bash
# Terminal
curl http://localhost:8090/health | jq

# Expected output:
{
  "status": "healthy",
  "timestamp": "2026-01-15T10:30:00Z"
}
```

**Screenshot 3**: Browser showing Fieldstone welcome page at localhost:8090

### 3. Authentication Flow

#### Register User
```bash
curl -X POST http://localhost:8090/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "demo@example.com", "password": "secure123"}' | jq
```

**Screenshot 4**: Terminal showing successful registration with user ID and token

#### Login
```bash
curl -X POST http://localhost:8090/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "demo@example.com", "password": "secure123"}' | jq
```

**Screenshot 5**: Login response showing JWT token

### 4. Collection Management

#### Create Collection
```bash
curl -X POST http://localhost:8090/api/collections \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "posts",
    "fields": [
      {"name": "title", "type": "text", "options": {"required": true}},
      {"name": "content", "type": "text"}
    ]
  }' | jq
```

**Screenshot 6**: Collection creation response

#### List Collections
```bash
curl http://localhost:8090/api/collections \
  -H "Authorization: Bearer <TOKEN>" | jq
```

**Screenshot 7**: Collections list showing posts collection

### 5. CRUD Operations

#### Create Record
```bash
curl -X POST http://localhost:8090/api/collections/<COLLECTION_ID>/records \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"title": "Hello World", "content": "My first post"}' | jq
```

**Screenshot 8**: Record creation with generated ID

#### Query Records
```bash
curl http://localhost:8090/api/collections/<COLLECTION_ID>/records \
  -H "Authorization: Bearer <TOKEN>" | jq
```

**Screenshot 9**: Paginated records list

### 6. Run Tests with Race Detection
```bash
go test -race ./internal/backend/... -v 2>&1 | head -50
```

**Screenshot 10**: Test output showing "PASS" and no race conditions

### 7. Benchmarks
```bash
go test -bench=. ./internal/backend/... -benchmem
```

**Screenshot 11**: Benchmark results showing operations per second

## Portfolio Page Structure

### Section 1: Project Overview
- **Title**: Fieldstone - Backend-as-a-Service in Go
- **Description**: PocketBase clone with multi-tenancy and PostgreSQL support
- **Tech Stack**: Go, SQLite, PostgreSQL, JWT, Chi Router
- **GitHub**: [Your repo link]

### Section 2: Architecture Diagram
Include screenshot of project structure:
```
fieldstone/
├── cmd/server/          # Main entry
├── internal/
│   ├── api/            # HTTP handlers
│   ├── auth/           # JWT authentication  
│   ├── backend/        # Storage abstraction
│   ├── jobs/           # Background workers ⭐
│   └── ratelimit/      # Rate limiting ⭐
```

### Section 3: Key Features
1. **Multi-threading**: Worker pools with goroutines
2. **Rate Limiting**: Token bucket algorithm
3. **Multi-tenancy**: Row-level security
4. **Dual Backend**: SQLite + PostgreSQL

### Section 4: Code Snippets
Show 3-4 key code blocks:
- Worker pool implementation
- Backend interface
- Rate limiting middleware

### Section 5: Test Results
Screenshots of:
- All tests passing
- Race detector clean
- Benchmark results

### Section 6: API Documentation
Screenshot of OpenAPI spec or example curl commands

## Screenshot Checklist

- [ ] Build success (terminal)
- [ ] Server running (terminal)
- [ ] Welcome page (browser)
- [ ] User registration (terminal)
- [ ] User login (terminal)
- [ ] Create collection (terminal)
- [ ] List collections (terminal)
- [ ] Create record (terminal)
- [ ] Query records (terminal)
- [ ] Tests passing (terminal)
- [ ] Race detector clean (terminal)
- [ ] VS Code workspace (IDE)
- [ ] Project structure (file explorer)

## Commands for Complete Demo

```bash
# Terminal 1: Start server
export PATH="/Users/Jre/go-install/go/bin:$PATH"
cd /Users/Jre/fieldstone
go run ./cmd/server

# Terminal 2: Run tests
export PATH="/Users/Jre/go-install/go/bin:$PATH"
cd /Users/Jre/fieldstone
go test -v ./...

# Terminal 3: API testing
export TOKEN=$(curl -s -X POST http://localhost:8090/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"secure123"}' | jq -r '.token')

echo "Token: $TOKEN"

# List collections
curl -s http://localhost:8090/api/collections -H "Authorization: Bearer $TOKEN" | jq
```

## Browser Testing

Open browser to:
1. `http://localhost:8090/` - Welcome page
2. `http://localhost:8090/health` - Health check JSON
3. Use REST Client extension in VS Code with provided .http files

## Video Demo Script (Optional)

1. **0:00-0:30**: Introduction, show project structure
2. **0:30-1:00**: Build and run server
3. **1:00-1:30**: Register and login
4. **1:30-2:00**: Create collection and add records
5. **2:00-2:30**: Run tests and show results
6. **2:30-3:00**: Show code highlights (worker pool, interfaces)

## Files to Highlight

1. `internal/jobs/queue.go` - Goroutine workers
2. `internal/backend/backend.go` - Interface design
3. `internal/api/ratelimit.go` - Rate limiting
4. `cmd/server/main.go` - Entry point
5. `docs/GO_TOPICS.md` - Documentation

## Tips for Portfolio

1. **Use dark theme** in terminal for better screenshots
2. **Use jq** for pretty JSON output
3. **Record GIF** usingasciinema or terminal recorder
4. **Add annotations** to screenshots
5. **Include code snippets** with syntax highlighting
6. **Show test coverage** percentage
7. **Mention lines of code**: ~3000+ Go code

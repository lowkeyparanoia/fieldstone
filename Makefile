.PHONY: help build test run dev clean install docker-build docker-run migrate lint fmt vet bench coverage

# Default target
help:
	@echo "Fieldstone - Open Source BaaS"
	@echo ""
	@echo "Available targets:"
	@echo "  make build          Build the server binary"
	@echo "  make build-all      Build for all platforms"
	@echo "  make test           Run all tests"
	@echo "  make test-coverage  Run tests with coverage report"
	@echo "  make test-race      Run tests with race detector"
	@echo "  make run            Run the server"
	@echo "  make dev            Run with hot reload (requires air)"
	@echo "  make clean          Remove build artifacts"
	@echo "  make install        Install to GOPATH/bin"
	@echo "  make docker-build   Build Docker image"
	@echo "  make docker-run     Run with Docker Compose"
	@echo "  make migrate        Run database migrations"
	@echo "  make migrate-down   Rollback migrations"
	@echo "  make migrate-status Show migration status"
	@echo "  make lint           Run linter (requires golangci-lint)"
	@echo "  make fmt            Format Go code"
	@echo "  make vet            Run go vet"
	@echo "  make bench          Run benchmarks"
	@echo "  make load-test      Run load tests (requires k6)"

# Build the server
build:
	@echo "Building Fieldstone server..."
	@mkdir -p bin
	go build -ldflags="-w -s" -o bin/fieldstone ./cmd/server
	go build -ldflags="-w -s" -o bin/fieldstone-migrate ./cmd/migrate
	@echo "Build complete: bin/fieldstone"

# Build for production (optimized)
build-prod:
	@echo "Building optimized binary..."
	@mkdir -p bin
	CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/fieldstone ./cmd/server
	@echo "Production build complete"

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p bin
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build -o bin/fieldstone-linux-amd64 ./cmd/server
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build -o bin/fieldstone-linux-arm64 ./cmd/server
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build -o bin/fieldstone-darwin-amd64 ./cmd/server
	# macOS ARM64
	GOOS=darwin GOARCH=arm64 go build -o bin/fieldstone-darwin-arm64 ./cmd/server
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build -o bin/fieldstone-windows-amd64.exe ./cmd/server
	@echo "All builds complete"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out
	@echo ""
	@echo "Coverage report generated: coverage.out"
	@echo "View HTML report: go tool cover -html=coverage.out -o coverage.html"

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	go test -race -v ./...

# Run the server
run: build
	./bin/fieldstone

# Run with hot reload (requires air)
dev:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Install to GOPATH/bin
install: build
	@echo "Installing to $(GOPATH)/bin..."
	@cp bin/fieldstone $(GOPATH)/bin/
	@cp bin/fieldstone-migrate $(GOPATH)/bin/
	@echo "Installation complete"

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t fieldstone:latest .

docker-run:
	@echo "Starting with Docker Compose..."
	docker-compose up -d

docker-stop:
	@echo "Stopping Docker Compose..."
	docker-compose down

docker-logs:
	docker-compose logs -f fieldstone

# Database migrations
migrate:
	@echo "Running migrations..."
	go run cmd/migrate/main.go -direction up

migrate-down:
	@echo "Rolling back migrations..."
	go run cmd/migrate/main.go -direction down -steps 1

migrate-status:
	@echo "Migration status:"
	go run cmd/migrate/main.go -status

# Linting and formatting
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run --timeout=5m

fmt:
	@echo "Formatting Go code..."
	gofmt -s -w .
	goimports -w .

vet:
	@echo "Running go vet..."
	go vet ./...

# Benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Load testing (requires k6)
load-test:
	@which k6 > /dev/null || (echo "Error: k6 not installed. Install from https://k6.io/docs/get-started/installation/" && exit 1)
	@echo "Running load tests..."
	k6 run tests/load/api_load_test.js

# Module management
mod-tidy:
	go mod tidy

mod-verify:
	go mod verify

mod-download:
	go mod download

# Frontend commands (admin dashboard)
frontend-install:
	cd web/admin && npm install

frontend-build:
	cd web/admin && npm run build

frontend-dev:
	cd web/admin && npm run dev

frontend-test:
	cd web/admin && npm test

frontend-lint:
	cd web/admin && npm run lint

# Development shortcuts for different backends
sqlite: build
	DATABASE_URL=sqlite://data/fieldstone.db ./bin/fieldstone

postgres: build
	DATABASE_URL=postgres://postgres:postgres@localhost:5432/fieldstone?sslmode=disable ./bin/fieldstone

# CI targets
ci: fmt vet test build

# Full release process
release: clean build-all test-coverage
	@echo "Release build complete. Binaries in bin/"

# Generate code (mocks, protobuf, etc.)
generate:
	go generate ./...

# Security scan
security-scan:
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	gosec ./...

# Dependency vulnerability check
vuln-check:
	@which govulncheck > /dev/null || (echo "Installing govulncheck..." && go install golang.org/x/vuln/cmd/govulncheck@latest)
	govulncheck ./...

# Documentation
docs-serve:
	@which mkdocs > /dev/null || (echo "Installing mkdocs..." && pip install mkdocs mkdocs-material)
	cd docs && mkdocs serve

docs-build:
	@which mkdocs > /dev/null || (echo "Installing mkdocs..." && pip install mkdocs mkdocs-material)
	cd docs && mkdocs build

# ---- WASM example plugin (slugify) ----
# Build the Rust -> wasm32-unknown-unknown example plugin.
plugin-build:
	@which rustc > /dev/null || (echo "install Rust: https://rustup.rs" && exit 1)
	rustup target add wasm32-unknown-unknown
	internal/plugins/examples/slugify/build.sh

# Run the example plugin against mock data via the real plugin runtime (needs Go 1.24).
plugin-test:
	go test ./internal/plugins -run TestSlugify -v

# Run the standalone harness (works on older Go; uses its own go.mod).
plugin-demo:
	cd internal/plugins/examples/slugify/harness && go run .

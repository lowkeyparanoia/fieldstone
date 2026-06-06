# Dockerfile for Fieldstone
# Multi-stage build for minimal production image

# Stage 1: Build
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=1 for SQLite support
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o fieldstone ./cmd/server

# Build migration tool
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o fieldstone-migrate ./cmd/migrate

# Stage 2: Frontend Build
FROM node:20-alpine AS frontend-builder

WORKDIR /app/web/admin
COPY web/admin/package*.json ./
RUN npm ci

COPY web/admin/ ./
RUN npm run build

# Stage 3: Production
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/fieldstone /usr/local/bin/
COPY --from=builder /app/fieldstone-migrate /usr/local/bin/

# Copy frontend build
COPY --from=frontend-builder /app/web/admin/dist ./web/admin/dist

# Create data directory
RUN mkdir -p /data && chmod 777 /data

# Expose ports
# 8090 - HTTP API and Admin UI
# 50051 - gRPC
EXPOSE 8090 50051

# Volume for persistent data
VOLUME ["/data"]

# Environment variables with defaults
ENV DATABASE_URL=sqlite:///data/fieldstone.db
ENV JWT_SECRET=change-me-in-production
ENV LOG_LEVEL=info
ENV ADMIN_UI_PATH=./web/admin/dist

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8090/health || exit 1

# Run migrations and start server
CMD fieldstone-migrate -dsn "$DATABASE_URL" -direction up && fieldstone

# Fieldstone

**Open Source Backend-as-a-Service (BaaS) - PocketBase Alternative**

[![CI](https://github.com/yourusername/fieldstone/actions/workflows/ci.yml/badge.svg)](https://github.com/yourusername/fieldstone/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/fieldstone)](https://goreportcard.com/report/github.com/yourusername/fieldstone)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker](https://img.shields.io/docker/pulls/yourusername/fieldstone)](https://hub.docker.com/r/yourusername/fieldstone)

Fieldstone is a lightweight, self-hosted backend platform that provides:
- **REST API** for CRUD operations
- **Real-time subscriptions** via WebSockets
- **Multi-tenancy** support
- **Authentication** with JWT
- **File storage** (local/S3)
- **Background jobs** with retry logic
- **Admin dashboard** for management
- **gRPC support** for high-performance operations

## 🚀 Quick Start

### Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/yourusername/fieldstone.git
cd fieldstone

# Start with Docker Compose
docker-compose up -d

# Access the services:
# API: http://localhost:8090
# Admin UI: http://localhost:8090/admin
# API Documentation: http://localhost:8090/docs
```

### Using Binary

```bash
# Download latest release
wget https://github.com/yourusername/fieldstone/releases/latest/download/fieldstone-linux-amd64.tar.gz
tar -xzf fieldstone-linux-amd64.tar.gz

# Run migrations
./fieldstone-migrate -dsn "data/fieldstone.db" -direction up

# Start server
./fieldstone
```

### Using Go

```bash
# Install
go install github.com/yourusername/fieldstone/cmd/server@latest

# Run
go run cmd/server/main.go
```

## 📋 Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Development](#development)
- [Deployment](#deployment)
- [Contributing](#contributing)
- [License](#license)

## ✨ Features

### Core Features

- ✅ **RESTful API** - Full CRUD operations for collections and records
- ✅ **Real-time** - WebSocket subscriptions for live updates
- ✅ **Authentication** - JWT-based auth with refresh tokens
- ✅ **Multi-tenancy** - Isolated data per tenant
- ✅ **File Storage** - Local filesystem or S3-compatible storage
- ✅ **Background Jobs** - Async job processing with retries
- ✅ **Rate Limiting** - Configurable rate limits per endpoint
- ✅ **Admin Dashboard** - React-based management interface
- ✅ **Database Support** - SQLite (dev) and PostgreSQL (prod)
- ✅ **Migrations** - Versioned database schema migrations
- ✅ **Webhooks** - Event-driven HTTP callbacks
- ✅ **gRPC** - High-performance RPC interface

### Security Features

- 🔒 JWT authentication with secure token rotation
- 🔒 Row-level security (RLS) with PostgreSQL
- 🔒 Password hashing with bcrypt
- 🔒 Rate limiting and request throttling
- 🔒 CORS configuration
- 🔒 Input validation and sanitization

### Developer Experience

- 🛠️ Auto-generated API documentation (OpenAPI/Swagger)
- 🛠️ TypeScript SDK for frontend integration
- 🛠️ Comprehensive test suite
- 🛠️ Hot reload for development
- 🛠️ Structured logging with zerolog
- 🛠️ Docker support with multi-stage builds

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         CLIENTS                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Web App     │  │  Mobile App  │  │  CLI Tool    │      │
│  │  (React)     │  │  (iOS/And)   │  │  (Go/TS)     │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
└─────────┼─────────────────┼─────────────────┼──────────────┘
          │                 │                 │
          └─────────────────┼─────────────────┘
                            │ HTTPS/HTTP2
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                       FIELDSTONE                             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  API Layer                                           │   │
│  │  • REST (Chi Router)    • gRPC    • WebSocket       │   │
│  └──────────────────┬───────────────────────────────────┘   │
│                     │                                        │
│  ┌──────────────────▼───────────────────────────────────┐   │
│  │  Service Layer                                       │   │
│  │  • Auth    • Collections    • Records    • Storage  │   │
│  └──────────────────┬───────────────────────────────────┘   │
│                     │                                        │
│  ┌──────────────────▼───────────────────────────────────┐   │
│  │  Backend Interface                                   │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────┐   │   │
│  │  │   SQLite     │  │  PostgreSQL  │  │  Memory  │   │   │
│  │  │   (dev)      │  │   (prod)     │  │  (test)  │   │   │
│  │  └──────────────┘  └──────────────┘  └──────────┘   │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## 🔧 Installation

### Prerequisites

- **Go 1.23+** (for building from source)
- **Node.js 20+** (for admin dashboard development)
- **Docker & Docker Compose** (optional, for containerized deployment)
- **PostgreSQL 15+** (optional, for production database)

### Docker Installation

```bash
# Pull the image
docker pull ghcr.io/yourusername/fieldstone:latest

# Run with SQLite (simplest)
docker run -d \
  --name fieldstone \
  -p 8090:8090 \
  -v fieldstone_data:/data \
  -e DATABASE_URL=sqlite:///data/fieldstone.db \
  ghcr.io/yourusername/fieldstone:latest

# Run with PostgreSQL (recommended for production)
docker-compose up -d
```

### Binary Installation

```bash
# Download for your platform
curl -L https://github.com/yourusername/fieldstone/releases/download/v1.0.0/fieldstone-$(uname -s)-$(uname -m).tar.gz | tar xz

# Move to PATH
sudo mv fieldstone /usr/local/bin/
sudo mv fieldstone-migrate /usr/local/bin/

# Verify installation
fieldstone --version
```

### Source Installation

```bash
# Clone repository
git clone https://github.com/yourusername/fieldstone.git
cd fieldstone

# Install Go dependencies
go mod download

# Build binary
make build

# Or build for production
make build-prod
```

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | Database connection string | `sqlite://data/fieldstone.db` |
| `JWT_SECRET` | Secret key for JWT signing | **Required** |
| `JWT_EXPIRY` | JWT token expiry duration | `1h` |
| `SERVER_HOST` | Server bind address | `0.0.0.0` |
| `SERVER_PORT` | HTTP server port | `8090` |
| `GRPC_PORT` | gRPC server port | `50051` |
| `LOG_LEVEL` | Logging level (debug/info/warn/error) | `info` |
| `STORAGE_TYPE` | Storage backend (local/s3) | `local` |
| `STORAGE_LOCAL_PATH` | Path for local file storage | `./uploads` |
| `RATE_LIMIT_ENABLED` | Enable rate limiting | `true` |
| `RATE_LIMIT_REQUESTS` | Requests per window | `100` |
| `RATE_LIMIT_WINDOW` | Rate limit window (seconds) | `60` |
| `ENABLE_WEBHOOKS` | Enable webhook processing | `true` |
| `ENABLE_PLUGINS` | Enable WASM plugins | `true` |
| `ENABLE_REALTIME` | Enable WebSocket realtime | `true` |

### Example Configuration

**Development (.env):**
```env
DATABASE_URL=sqlite://data/fieldstone.db
JWT_SECRET=dev-secret-not-for-production
LOG_LEVEL=debug
SERVER_PORT=8090
STORAGE_TYPE=local
STORAGE_LOCAL_PATH=./uploads
```

**Production (.env.production):**
```env
DATABASE_URL=postgres://user:password@postgres:5432/fieldstone?sslmode=require
JWT_SECRET=your-256-bit-secret-key-here
LOG_LEVEL=warn
SERVER_PORT=8090
STORAGE_TYPE=s3
STORAGE_S3_ENDPOINT=s3.amazonaws.com
STORAGE_S3_BUCKET=fieldstone-files
STORAGE_S3_REGION=us-east-1
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=1000
```

## 📚 API Documentation

### Authentication

**Register a new user:**
```bash
curl -X POST http://localhost:8090/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8090/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123"
  }'
# Response: {"token": "eyJhbGciOiJIUzI1NiIs...", "record": {...}}
```

**Use the token:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8090/api/collections
```

### Collections

**Create a collection:**
```bash
curl -X POST http://localhost:8090/api/collections \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "products",
    "schema": [
      {"name": "title", "type": "text", "required": true},
      {"name": "price", "type": "number"},
      {"name": "inStock", "type": "boolean", "default": true}
    ]
  }'
```

**List collections:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8090/api/collections
```

### Records

**Create a record:**
```bash
curl -X POST http://localhost:8090/api/collections/PRODUCTS_ID/records \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Awesome Product",
    "price": 99.99,
    "inStock": true
  }'
```

**Query records with pagination:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8090/api/collections/PRODUCTS_ID/records?page=1&perPage=10&sort=-created"
```

### Real-time Subscriptions (WebSocket)

Connect to WebSocket:
```javascript
const ws = new WebSocket('ws://localhost:8090/ws');

// Subscribe to collection
ws.send(JSON.stringify({
  type: 'subscribe',
  collectionId: 'PRODUCTS_ID'
}));

// Listen for events
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Change detected:', data);
};
```

## 🛠️ Development

### Setup Development Environment

```bash
# Clone repo
git clone https://github.com/yourusername/fieldstone.git
cd fieldstone

# Install Go dependencies
go mod download

# Install frontend dependencies
cd web/admin && npm install && cd ../..

# Copy environment file
cp .env.example .env
# Edit .env with your settings

# Run database migrations
go run cmd/migrate/main.go -direction up

# Start development server with hot reload
make dev
```

### Project Structure

```
fieldstone/
├── cmd/                    # Application entry points
│   ├── server/            # Main HTTP server
│   └── migrate/           # Database migration tool
├── internal/              # Private application code
│   ├── api/              # HTTP handlers
│   ├── auth/             # Authentication logic
│   ├── backend/          # Database implementations
│   ├── jobs/             # Background job queue
│   ├── migrate/          # Migration framework
│   ├── ratelimit/        # Rate limiting
│   ├── realtime/         # WebSocket manager
│   ├── storage/          # File storage
│   └── webhooks/         # Webhook processing
├── pkg/                   # Public packages
│   └── models/           # Data models
├── web/                   # Frontend applications
│   └── admin/            # React admin dashboard
├── api/                   # API definitions
│   └── proto/            # gRPC protobuf files
├── docs/                  # Documentation
├── migrations/            # Database migrations
├── scripts/               # Utility scripts
├── Dockerfile            # Docker build file
├── docker-compose.yml    # Docker Compose config
├── Makefile              # Build automation
└── README.md             # This file
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package tests
go test ./internal/api/... -v

# Run benchmarks
go test ./... -bench=.
```

### Adding Migrations

```bash
# Create new migration files
# SQLite:
touch internal/migrate/migrations/sqlite/002_add_users_field.up.sql
touch internal/migrate/migrations/sqlite/002_add_users_field.down.sql

# PostgreSQL:
touch internal/migrate/migrations/postgres/002_add_users_field.up.sql
touch internal/migrate/migrations/postgres/002_add_users_field.down.sql

# Write migration SQL, then apply:
go run cmd/migrate/main.go -direction up
```

## 🚀 Deployment

### Docker Compose (Recommended)

```bash
# Production deployment
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# View logs
docker-compose logs -f

# Scale API servers
docker-compose up -d --scale fieldstone=3
```

### Kubernetes

```bash
# Apply manifests
kubectl apply -f k8s/

# Check deployment
kubectl get pods -l app=fieldstone

# View logs
kubectl logs -l app=fieldstone -f
```

### Systemd Service

Create `/etc/systemd/system/fieldstone.service`:

```ini
[Unit]
Description=Fieldstone BaaS
After=network.target

[Service]
Type=simple
User=fieldstone
WorkingDirectory=/opt/fieldstone
ExecStart=/usr/local/bin/fieldstone
Restart=always
RestartSec=10
Environment="DATABASE_URL=sqlite:///data/fieldstone.db"
Environment="JWT_SECRET=your-secret-key"

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable fieldstone
sudo systemctl start fieldstone
sudo systemctl status fieldstone
```

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Workflow

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Code Standards

- Follow standard Go formatting (`gofmt`)
- Write comprehensive tests
- Maintain 70%+ code coverage
- Update documentation
- Follow conventional commits

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by [PocketBase](https://pocketbase.io/)
- Built with [Chi Router](https://go-chi.io/)
- UI components from [shadcn/ui](https://ui.shadcn.com/)

## 📞 Support

- **Documentation**: [https://fieldstone-docs.vercel.app](https://fieldstone-docs.vercel.app)
- **Issues**: [GitHub Issues](https://github.com/yourusername/fieldstone/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/fieldstone/discussions)
- **Discord**: [Join our community](https://discord.gg/fieldstone)

---

**Made with ❤️ by the Fieldstone community**

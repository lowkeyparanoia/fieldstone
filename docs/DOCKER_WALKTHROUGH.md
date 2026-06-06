# Docker & Containerization Walkthrough for Fieldstone

## What is Docker?

**Docker** is a platform that allows you to package applications with all their dependencies into standardized units called **containers**. Think of it like shipping containers in the logistics industry - just as a shipping container packages up goods so they can be moved anywhere in the world, a Docker container packages up software so it can run anywhere.

### Why Use Docker?

**The Problem (Before Docker):**
```
Developer A: "It works on my machine!"
Developer B: "But it doesn't work on mine..."
Server: "I have different library versions!"
```

**The Solution (With Docker):**
```
Everyone: "It works everywhere because we all use the same container!"
```

### Key Concepts

#### 1. **Containers**
A container is a lightweight, standalone, executable package that includes:
- Your application code
- Runtime environment
- System tools
- System libraries
- Settings

**Analogy:** Think of a container like an apartment in a building:
- Each apartment (container) has its own isolated space
- They share the building's infrastructure (OS kernel)
- Each can have different furniture (dependencies) without affecting others

#### 2. **Images**
An image is a read-only template used to create containers. It's like a blueprint or a recipe.

**Analogy:** An image is like a cookie cutter:
- The cookie cutter (image) defines the shape
- You can make many cookies (containers) from one cutter
- All cookies will be identical

#### 3. **Dockerfile**
A Dockerfile is a text document that contains instructions to build an image.

**Example:** It's like a recipe that tells Docker:
1. Start with Go 1.23
2. Add the source code
3. Compile the application
4. Set up the environment

#### 4. **Volumes**
Volumes are directories that exist outside the container's filesystem, used for persistent data.

**Analogy:** Like a USB drive:
- The container is your computer
- The volume is a USB drive plugged in
- If you replace the computer (container), the USB drive (data) remains

#### 5. **Networking**
Docker creates isolated networks for containers to communicate.

**Analogy:** Like a private office network:
- Containers are computers in the office
- They can talk to each other on the internal network
- Specific ports can be exposed to the outside world

---

## How Docker Works with Fieldstone

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        YOUR COMPUTER                         │
├─────────────────────────────────────────────────────────────┤
│  Docker Engine                                               │
│  ┌─────────────────────────────────────────────────────────┐│
│  │  Fieldstone Container                                    ││
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ││
│  │  │  Fieldstone  │  │  Admin UI    │  │  Migrations  │  ││
│  │  │  Server      │  │  (React)     │  │  Tool        │  ││
│  │  │  Port: 8090  │  │              │  │              │  ││
│  │  └──────────────┘  └──────────────┘  └──────────────┘  ││
│  │                                                         ││
│  │  Volume: /data ────────┐                               ││
│  └────────────────────────┼───────────────────────────────┘│
│                           │                                 │
│  ┌────────────────────────┼───────────────────────────────┐│
│  │  PostgreSQL Container  │                               ││
│  │  Port: 5432            │                               ││
│  │                        │                               ││
│  │  Volume: /var/lib/...─┘                               ││
│  └─────────────────────────────────────────────────────────┘│
│                           │                                 │
│  ┌────────────────────────┘                                 │
│  │  Redis Container                                          │
│  │  Port: 6379                                               │
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

### Multi-Stage Build Explained

Fieldstone's Dockerfile uses **multi-stage builds** for efficiency:

```dockerfile
# Stage 1: Builder (Heavy, has all build tools)
FROM golang:1.23-alpine AS builder
# ... compile the application ...

# Stage 2: Frontend Builder (Node.js environment)
FROM node:20-alpine AS frontend-builder  
# ... build React app ...

# Stage 3: Production (Light, only runtime)
FROM alpine:latest
# ... copy compiled binaries ...
# ... copy built frontend ...
```

**Why Multi-Stage?**

| Stage | Size | Contains |
|-------|------|----------|
| Builder | ~1GB | Go compiler, build tools, source code |
| Frontend | ~500MB | Node.js, npm, build tools |
| **Production** | **~50MB** | **Only compiled binary + runtime libs** |

**Benefits:**
- ✅ Smaller final image (50MB vs 1.5GB)
- ✅ Faster deployments
- ✅ Less attack surface (security)
- ✅ No build tools in production

---

## Step-by-Step Docker Usage

### Step 1: Install Docker

**macOS:**
```bash
# Using Homebrew
brew install --cask docker

# Or download from https://docs.docker.com/desktop/install/mac-install/
```

**Linux (Ubuntu/Debian):**
```bash
# Update package index
sudo apt update

# Install prerequisites
sudo apt install apt-transport-https ca-certificates curl gnupg lsb-release

# Add Docker's official GPG key
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# Set up the stable repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine
sudo apt update
sudo apt install docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Verify installation
sudo docker run hello-world
```

**Windows:**
Download Docker Desktop from https://docs.docker.com/desktop/install/windows-install/

### Step 2: Build Fieldstone Image

```bash
# Navigate to Fieldstone directory
cd /path/to/fieldstone

# Build the Docker image
# Format: docker build -t <name>:<tag> .
docker build -t fieldstone:latest .

# This command:
# -t fieldstone:latest  → Names the image "fieldstone" with tag "latest"
# .                     → Uses Dockerfile in current directory

# Verify image was created
docker images | grep fieldstone

# Output:
# REPOSITORY    TAG       IMAGE ID       CREATED         SIZE
# fieldstone    latest    abc123def456   2 minutes ago   52MB
```

**What Happens During Build:**

1. **Download base images** (Go, Node, Alpine)
2. **Install dependencies** (go mod download, npm install)
3. **Copy source code**
4. **Compile application** (go build)
5. **Build frontend** (npm run build)
6. **Create minimal production image**

### Step 3: Run Fieldstone Container

**Basic Run (SQLite):**
```bash
# Run with SQLite
docker run -d \
  --name fieldstone \
  -p 8090:8090 \
  -p 50051:50051 \
  -v fieldstone_data:/data \
  -e DATABASE_URL=sqlite:///data/fieldstone.db \
  -e JWT_SECRET=your-secret-key \
  fieldstone:latest

# Flags explained:
# -d                    → Run in background (detached mode)
# --name fieldstone     → Name the container
# -p 8090:8090          → Map host port 8090 to container port 8090
# -p 50051:50051        → Map gRPC port
# -v fieldstone_data:/data  → Create persistent volume
# -e KEY=value          → Set environment variables
```

**With PostgreSQL (Full Stack):**
```bash
# Using Docker Compose (recommended)
docker-compose up -d

# This starts:
# - Fieldstone API server
# - PostgreSQL database
# - Redis cache
# - MinIO object storage

# View logs
docker-compose logs -f fieldstone

# Stop everything
docker-compose down

# Stop and remove volumes (WARNING: deletes data!)
docker-compose down -v
```

### Step 4: Manage Containers

```bash
# List running containers
docker ps

# List all containers (including stopped)
docker ps -a

# View container logs
docker logs fieldstone
docker logs -f fieldstone  # Follow mode

# Execute commands inside container
docker exec -it fieldstone sh

# Inside container, you can:
# - Run migrations: fieldstone-migrate -direction up
# - Check files: ls -la /data
# - View processes: ps aux

# Stop container
docker stop fieldstone

# Start stopped container
docker start fieldstone

# Restart container
docker restart fieldstone

# Remove container
docker rm fieldstone

# Remove container and its volumes
docker rm -v fieldstone
```

### Step 5: Data Persistence

**Understanding Volumes:**

```bash
# Create a named volume
docker volume create fieldstone_data

# Run container with volume
docker run -d \
  --name fieldstone \
  -v fieldstone_data:/data \
  fieldstone:latest

# What's happening:
# Host: /var/lib/docker/volumes/fieldstone_data/_data
# Container: /data
# These two locations are synchronized!

# Backup data
docker run --rm \
  -v fieldstone_data:/source \
  -v $(pwd):/backup \
  alpine tar czf /backup/fieldstone-backup.tar.gz -C /source .

# Restore data
docker run --rm \
  -v fieldstone_data:/target \
  -v $(pwd):/backup \
  alpine sh -c "cd /target && tar xzf /backup/fieldstone-backup.tar.gz"
```

### Step 6: Environment Configuration

**Development Environment:**
```bash
# Create .env file
cat > .env << EOF
DATABASE_URL=sqlite:///data/fieldstone.db
JWT_SECRET=dev-secret-do-not-use-in-production
LOG_LEVEL=debug
ENABLE_DEBUG=true
EOF

# Run with environment file
docker run -d --env-file .env fieldstone:latest
```

**Production Environment:**
```bash
# Secure production setup
cat > .env.production << EOF
DATABASE_URL=postgres://user:password@postgres:5432/fieldstone
JWT_SECRET=$(openssl rand -hex 32)
LOG_LEVEL=warn
RATE_LIMIT_ENABLED=true
STORAGE_TYPE=s3
STORAGE_S3_BUCKET=fieldstone-files
EOF

docker run -d \
  --env-file .env.production \
  --restart unless-stopped \
  fieldstone:latest
```

---

## Docker Compose Deep Dive

### What is Docker Compose?

Docker Compose is a tool for defining and running **multi-container applications**. Instead of running multiple `docker run` commands, you define everything in a `docker-compose.yml` file.

**Analogy:** If Docker containers are musical instruments, Docker Compose is the conductor orchestrating them together.

### Fieldstone's Docker Compose Structure

```yaml
version: '3.8'

services:
  # Each service is a container
  fieldstone:     # Main application
  postgres:       # Database
  redis:          # Cache
  minio:          # File storage
  nginx:          # Reverse proxy (optional)

volumes:
  # Persistent storage
  fieldstone_data:
  postgres_data:

networks:
  # Internal communication
  fieldstone-network:
```

### Common Docker Compose Commands

```bash
# Start all services
docker-compose up -d

# Start specific services
docker-compose up -d fieldstone postgres

# View logs
docker-compose logs
docker-compose logs -f fieldstone  # Follow fieldstone only

# Scale services
docker-compose up -d --scale fieldstone=3

# Build images before starting
docker-compose up -d --build

# Run migrations
docker-compose exec fieldstone fieldstone-migrate -direction up

# Execute command in service
docker-compose exec postgres psql -U postgres -d fieldstone

# View resource usage
docker-compose stats

# Stop services
docker-compose stop

# Remove containers
docker-compose down

# Remove everything including volumes
docker-compose down -v
```

---

## Production Deployment Guide

### 1. Build Production Image

```bash
# Tag with version
docker build -t fieldstone:v1.0.0 .

# Push to registry (Docker Hub, GitHub Container Registry, etc.)
docker tag fieldstone:v1.0.0 ghcr.io/yourusername/fieldstone:v1.0.0
docker push ghcr.io/yourusername/fieldstone:v1.0.0
```

### 2. Docker Swarm (Simple Orchestration)

```bash
# Initialize swarm
docker swarm init

# Deploy stack
docker stack deploy -c docker-compose.yml fieldstone

# View services
docker stack services fieldstone

# Scale service
docker service scale fieldstone_fieldstone=3
```

### 3. Kubernetes Deployment

```yaml
# kubernetes-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fieldstone
spec:
  replicas: 3
  selector:
    matchLabels:
      app: fieldstone
  template:
    metadata:
      labels:
        app: fieldstone
    spec:
      containers:
      - name: fieldstone
        image: fieldstone:latest
        ports:
        - containerPort: 8090
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: fieldstone-secrets
              key: database-url
```

Deploy:
```bash
kubectl apply -f kubernetes-deployment.yaml
```

---

## Troubleshooting Common Issues

### Issue 1: Port Already in Use
```bash
# Error: bind: address already in use
# Solution: Find and kill process or use different port
docker run -p 8091:8090 fieldstone:latest  # Use port 8091 on host
```

### Issue 2: Permission Denied
```bash
# Error: permission denied while trying to connect to Docker daemon
# Solution: Add user to docker group
sudo usermod -aG docker $USER
# Log out and back in
```

### Issue 3: Container Exits Immediately
```bash
# Check logs
docker logs fieldstone

# Run interactively to see errors
docker run -it --rm fieldstone:latest sh
```

### Issue 4: Database Connection Failed
```bash
# Ensure database container is running
docker-compose ps

# Check network connectivity
docker-compose exec fieldstone ping postgres

# Verify environment variables
docker-compose exec fieldstone env | grep DATABASE
```

### Issue 5: Image Build Fails
```bash
# Clear build cache
docker build --no-cache -t fieldstone:latest .

# Check available disk space
docker system df

# Clean up unused resources
docker system prune -a
```

---

## Best Practices

### 1. **Security**
- ✅ Never hardcode secrets in Dockerfile
- ✅ Use multi-stage builds to minimize attack surface
- ✅ Run containers as non-root user
- ✅ Keep base images updated
- ✅ Scan images for vulnerabilities: `docker scan fieldstone`

### 2. **Performance**
- ✅ Use .dockerignore to exclude unnecessary files
- ✅ Order Dockerfile instructions by change frequency
- ✅ Combine RUN commands to reduce layers
- ✅ Use specific image tags (not `latest`)

### 3. **Maintenance**
- ✅ Use semantic versioning for images
- ✅ Keep docker-compose.yml in version control
- ✅ Document environment variables
- ✅ Use health checks
- ✅ Set resource limits

### 4. **Development Workflow**
```bash
# 1. Make code changes
# 2. Build image
docker-compose build

# 3. Run tests
docker-compose run --rm fieldstone go test ./...

# 4. Start services
docker-compose up -d

# 5. Verify
curl http://localhost:8090/health
```

---

## Quick Reference Card

```bash
# BUILD
docker build -t fieldstone:latest .

# RUN
docker run -d -p 8090:8090 -v data:/data fieldstone:latest

# COMPOSE
docker-compose up -d
docker-compose down
docker-compose logs -f

# DEBUG
docker ps
docker logs container
docker exec -it container sh

# CLEANUP
docker system prune        # Remove unused data
docker volume prune        # Remove unused volumes
docker image prune         # Remove unused images
```

---

**Next Steps:** After understanding Docker, proceed to set up CI/CD with GitHub Actions to automatically build and test your Docker images on every code change.

---
applyTo: 'Dockerfile,docker-compose.yml,.air.toml,.env*,config/**/*'
---

# Configuration and Development Tools

This document covers Docker configuration, development tooling setup, environment management, and configuration best practices for the Go Chat Widget Dashboard.

## Docker Configuration

### Multi-Stage Production Dockerfile
```dockerfile
# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o server ./cmd/server

# Final stage - minimal runtime image
FROM scratch

# Copy timezone data and CA certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy application binary
COPY --from=builder /app/server /server

# Copy static assets
COPY --from=builder /app/web/static /web/static
COPY --from=builder /app/web/templates /web/templates

# Create directories for uploads and logs
USER 1000:1000

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/server", "health"]

EXPOSE 8080

ENTRYPOINT ["/server"]
```

### Development Docker Compose
```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    ports:
      - "8080:8080"
      - "2345:2345"  # Delve debugger port
    environment:
      - DATABASE_URL=postgres://chatuser:chatpass@db:5432/chatwidget?sslmode=disable
      - REDIS_URL=redis://redis:6379/0
      - LOG_LEVEL=debug
      - ENV=development
      - ENABLE_DELVE=true
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - .:/app
      - go_cache:/go/pkg/mod
      - ./uploads:/app/uploads
    restart: unless-stopped
    networks:
      - chat-network

  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: chatwidget
      POSTGRES_USER: chatuser
      POSTGRES_PASSWORD: chatpass
      POSTGRES_INITDB_ARGS: "--auth-host=md5"
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/01-init.sql
      - ./scripts/seed-data.sql:/docker-entrypoint-initdb.d/02-seed.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U chatuser -d chatwidget"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - chat-network

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
      - ./config/redis.conf:/usr/local/etc/redis/redis.conf
    command: redis-server /usr/local/etc/redis/redis.conf
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5
    networks:
      - chat-network

  # Development tools
  adminer:
    image: adminer:4.8.1
    ports:
      - "8081:8080"
    environment:
      ADMINER_DEFAULT_SERVER: db
    depends_on:
      - db
    networks:
      - chat-network

  redis-commander:
    image: rediscommander/redis-commander:latest
    ports:
      - "8082:8081"
    environment:
      REDIS_HOSTS: local:redis:6379
    depends_on:
      - redis
    networks:
      - chat-network

  # Monitoring stack (optional for development)
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./config/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--web.enable-lifecycle'
    networks:
      - chat-network

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin
    volumes:
      - grafana_data:/var/lib/grafana
      - ./config/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./config/grafana/datasources:/etc/grafana/provisioning/datasources
    depends_on:
      - prometheus
    networks:
      - chat-network

volumes:
  postgres_data:
  redis_data:
  go_cache:
  prometheus_data:
  grafana_data:

networks:
  chat-network:
    driver: bridge
```

### Development Dockerfile
```dockerfile
# Dockerfile.dev
FROM golang:1.21-alpine

# Install development tools
RUN apk add --no-cache git ca-certificates tzdata curl

# Install delve for debugging
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Install air for hot reloading
RUN go install github.com/cosmtrek/air@latest

# Install other development tools
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
RUN go install golang.org/x/tools/cmd/goimports@latest

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Expose application and debugger ports
EXPOSE 8080 2345

# Use air for hot reloading in development
CMD ["air"]
```

## Air Configuration (Hot Reloading)

### .air.toml
```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/server"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "web/node_modules"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html", "css", "js"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = true

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

## Environment Configuration

### Environment Variables Management
```bash
# .env.example - Template for environment variables
# Copy this to .env and update values for local development

# Application Configuration
PORT=8080
ENV=development
LOG_LEVEL=debug
APP_NAME="Chat Widget Dashboard"
APP_VERSION=1.0.0

# Database Configuration
DATABASE_URL=postgres://chatuser:chatpass@localhost:5432/chatwidget?sslmode=disable
DATABASE_MAX_OPEN_CONNS=25
DATABASE_MAX_IDLE_CONNS=5
DATABASE_CONN_MAX_LIFETIME=1h

# Redis Configuration
REDIS_URL=redis://localhost:6379/0
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10

# Authentication & Security
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_EXPIRES_IN=24h
CSRF_SECRET=your-csrf-secret-key
SESSION_SECRET=your-session-secret-key

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With

# Email Configuration (Optional)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
EMAIL_FROM=noreply@yourcompany.com

# External Services
OPENAI_API_KEY=your-openai-api-key
WEBHOOK_SECRET=your-webhook-secret

# File Upload Configuration
UPLOAD_MAX_SIZE=10MB
UPLOAD_ALLOWED_TYPES=image/jpeg,image/png,image/gif,application/pdf

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# WebSocket Configuration
WS_READ_BUFFER_SIZE=1024
WS_WRITE_BUFFER_SIZE=1024
WS_MAX_MESSAGE_SIZE=512

# Monitoring & Observability
ENABLE_METRICS=true
METRICS_PATH=/metrics
HEALTH_CHECK_PATH=/health

# Development Only
ENABLE_DELVE=false
DELVE_PORT=2345
ENABLE_PPROF=true
PPROF_PORT=6060
```

### Configuration Loading
```go
// config/config.go
package config

import (
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/joho/godotenv"
)

type Config struct {
    // Application
    Port        string        `env:"PORT" default:"8080"`
    Environment string        `env:"ENV" default:"development"`
    LogLevel    string        `env:"LOG_LEVEL" default:"info"`
    AppName     string        `env:"APP_NAME" default:"Chat Widget Dashboard"`
    Version     string        `env:"APP_VERSION" default:"1.0.0"`

    // Database
    DatabaseURL             string        `env:"DATABASE_URL" required:"true"`
    DatabaseMaxOpenConns    int           `env:"DATABASE_MAX_OPEN_CONNS" default:"25"`
    DatabaseMaxIdleConns    int           `env:"DATABASE_MAX_IDLE_CONNS" default:"5"`
    DatabaseConnMaxLifetime time.Duration `env:"DATABASE_CONN_MAX_LIFETIME" default:"1h"`

    // Redis
    RedisURL      string `env:"REDIS_URL" required:"true"`
    RedisPassword string `env:"REDIS_PASSWORD"`
    RedisDB       int    `env:"REDIS_DB" default:"0"`
    RedisPoolSize int    `env:"REDIS_POOL_SIZE" default:"10"`

    // Security
    JWTSecret     string        `env:"JWT_SECRET" required:"true"`
    JWTExpiresIn  time.Duration `env:"JWT_EXPIRES_IN" default:"24h"`
    CSRFSecret    string        `env:"CSRF_SECRET" required:"true"`
    SessionSecret string        `env:"SESSION_SECRET" required:"true"`

    // CORS
    CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS"`
    CORSAllowedMethods []string `env:"CORS_ALLOWED_METHODS"`
    CORSAllowedHeaders []string `env:"CORS_ALLOWED_HEADERS"`

    // Email
    SMTPHost     string `env:"SMTP_HOST"`
    SMTPPort     int    `env:"SMTP_PORT" default:"587"`
    SMTPUsername string `env:"SMTP_USERNAME"`
    SMTPPassword string `env:"SMTP_PASSWORD"`
    EmailFrom    string `env:"EMAIL_FROM"`

    // External Services
    OpenAIAPIKey  string `env:"OPENAI_API_KEY"`
    WebhookSecret string `env:"WEBHOOK_SECRET"`

    // File Upload
    UploadMaxSize      string   `env:"UPLOAD_MAX_SIZE" default:"10MB"`
    UploadAllowedTypes []string `env:"UPLOAD_ALLOWED_TYPES"`

    // Rate Limiting
    RateLimitRequests int           `env:"RATE_LIMIT_REQUESTS" default:"100"`
    RateLimitWindow   time.Duration `env:"RATE_LIMIT_WINDOW" default:"1m"`

    // WebSocket
    WSReadBufferSize   int `env:"WS_READ_BUFFER_SIZE" default:"1024"`
    WSWriteBufferSize  int `env:"WS_WRITE_BUFFER_SIZE" default:"1024"`
    WSMaxMessageSize   int `env:"WS_MAX_MESSAGE_SIZE" default:"512"`

    // Monitoring
    EnableMetrics     bool   `env:"ENABLE_METRICS" default:"true"`
    MetricsPath       string `env:"METRICS_PATH" default:"/metrics"`
    HealthCheckPath   string `env:"HEALTH_CHECK_PATH" default:"/health"`

    // Development
    EnableDelve bool `env:"ENABLE_DELVE" default:"false"`
    DelvePort   int  `env:"DELVE_PORT" default:"2345"`
    EnablePprof bool `env:"ENABLE_PPROF" default:"false"`
    PprofPort   int  `env:"PPROF_PORT" default:"6060"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
    // Load .env file if it exists (development)
    if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
        return nil, fmt.Errorf("error loading .env file: %w", err)
    }

    config := &Config{}

    // Load configuration using reflection or manual parsing
    if err := loadFromEnv(config); err != nil {
        return nil, fmt.Errorf("error loading configuration: %w", err)
    }

    // Validate required fields
    if err := validate(config); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }

    return config, nil
}

func loadFromEnv(config *Config) error {
    // Application
    config.Port = getEnvOrDefault("PORT", "8080")
    config.Environment = getEnvOrDefault("ENV", "development")
    config.LogLevel = getEnvOrDefault("LOG_LEVEL", "info")
    config.AppName = getEnvOrDefault("APP_NAME", "Chat Widget Dashboard")
    config.Version = getEnvOrDefault("APP_VERSION", "1.0.0")

    // Database
    config.DatabaseURL = os.Getenv("DATABASE_URL")
    config.DatabaseMaxOpenConns = getEnvAsInt("DATABASE_MAX_OPEN_CONNS", 25)
    config.DatabaseMaxIdleConns = getEnvAsInt("DATABASE_MAX_IDLE_CONNS", 5)
    config.DatabaseConnMaxLifetime = getEnvAsDuration("DATABASE_CONN_MAX_LIFETIME", time.Hour)

    // Redis
    config.RedisURL = os.Getenv("REDIS_URL")
    config.RedisPassword = os.Getenv("REDIS_PASSWORD")
    config.RedisDB = getEnvAsInt("REDIS_DB", 0)
    config.RedisPoolSize = getEnvAsInt("REDIS_POOL_SIZE", 10)

    // Security
    config.JWTSecret = os.Getenv("JWT_SECRET")
    config.JWTExpiresIn = getEnvAsDuration("JWT_EXPIRES_IN", 24*time.Hour)
    config.CSRFSecret = os.Getenv("CSRF_SECRET")
    config.SessionSecret = os.Getenv("SESSION_SECRET")

    // CORS
    config.CORSAllowedOrigins = getEnvAsSlice("CORS_ALLOWED_ORIGINS", ",")
    config.CORSAllowedMethods = getEnvAsSlice("CORS_ALLOWED_METHODS", ",")
    config.CORSAllowedHeaders = getEnvAsSlice("CORS_ALLOWED_HEADERS", ",")

    // Email
    config.SMTPHost = os.Getenv("SMTP_HOST")
    config.SMTPPort = getEnvAsInt("SMTP_PORT", 587)
    config.SMTPUsername = os.Getenv("SMTP_USERNAME")
    config.SMTPPassword = os.Getenv("SMTP_PASSWORD")
    config.EmailFrom = os.Getenv("EMAIL_FROM")

    // External Services
    config.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
    config.WebhookSecret = os.Getenv("WEBHOOK_SECRET")

    // File Upload
    config.UploadMaxSize = getEnvOrDefault("UPLOAD_MAX_SIZE", "10MB")
    config.UploadAllowedTypes = getEnvAsSlice("UPLOAD_ALLOWED_TYPES", ",")

    // Rate Limiting
    config.RateLimitRequests = getEnvAsInt("RATE_LIMIT_REQUESTS", 100)
    config.RateLimitWindow = getEnvAsDuration("RATE_LIMIT_WINDOW", time.Minute)

    // WebSocket
    config.WSReadBufferSize = getEnvAsInt("WS_READ_BUFFER_SIZE", 1024)
    config.WSWriteBufferSize = getEnvAsInt("WS_WRITE_BUFFER_SIZE", 1024)
    config.WSMaxMessageSize = getEnvAsInt("WS_MAX_MESSAGE_SIZE", 512)

    // Monitoring
    config.EnableMetrics = getEnvAsBool("ENABLE_METRICS", true)
    config.MetricsPath = getEnvOrDefault("METRICS_PATH", "/metrics")
    config.HealthCheckPath = getEnvOrDefault("HEALTH_CHECK_PATH", "/health")

    // Development
    config.EnableDelve = getEnvAsBool("ENABLE_DELVE", false)
    config.DelvePort = getEnvAsInt("DELVE_PORT", 2345)
    config.EnablePprof = getEnvAsBool("ENABLE_PPROF", false)
    config.PprofPort = getEnvAsInt("PPROF_PORT", 6060)

    return nil
}

func validate(config *Config) error {
    required := map[string]string{
        "DATABASE_URL":    config.DatabaseURL,
        "REDIS_URL":       config.RedisURL,
        "JWT_SECRET":      config.JWTSecret,
        "CSRF_SECRET":     config.CSRFSecret,
        "SESSION_SECRET":  config.SessionSecret,
    }

    for field, value := range required {
        if value == "" {
            return fmt.Errorf("required environment variable %s is not set", field)
        }
    }

    // Validate environment
    validEnvs := []string{"development", "staging", "production"}
    if !contains(validEnvs, config.Environment) {
        return fmt.Errorf("invalid environment: %s. Must be one of: %s", 
            config.Environment, strings.Join(validEnvs, ", "))
    }

    return nil
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if boolValue, err := strconv.ParseBool(value); err == nil {
            return boolValue
        }
    }
    return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}

func getEnvAsSlice(key, separator string) []string {
    if value := os.Getenv(key); value != "" {
        return strings.Split(value, separator)
    }
    return []string{}
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
    return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
    return c.Environment == "production"
}
```

## Database Configuration

### Database Initialization Script
```sql
-- scripts/init-db.sql
-- Database initialization script

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- Create custom types
DO $ BEGIN
    CREATE TYPE user_role AS ENUM ('admin', 'agent', 'customer');
EXCEPTION
    WHEN duplicate_object THEN null;
END $;

DO $ BEGIN
    CREATE TYPE message_type AS ENUM ('text', 'image', 'file', 'system');
EXCEPTION
    WHEN duplicate_object THEN null;
END $;

-- Create sequences
CREATE SEQUENCE IF NOT EXISTS global_id_seq;

-- Create functions
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$ language 'plpgsql';

-- Create indexes after tables are created by GORM
-- These will be created in migration files
```

### Seed Data Script
```sql
-- scripts/seed-data.sql
-- Development seed data

-- Insert admin user
INSERT INTO users (id, name, email, password, role, created_at, updated_at) 
VALUES (
    1,
    'Admin User',
    'admin@example.com',
    '$2a$10$8K1p/a0gDt2FGnKaGzKGC.UMYx3KqK5DZ3FZh1Kk7U9f6h7.mQ2.m', -- password: admin123
    'admin',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
) ON CONFLICT (email) DO NOTHING;

-- Insert test customers
INSERT INTO users (id, name, email, password, role, created_at, updated_at) 
VALUES 
    (2, 'John Doe', 'john@example.com', '$2a$10$8K1p/a0gDt2FGnKaGzKGC.UMYx3KqK5DZ3FZh1Kk7U9f6h7.mQ2.m', 'customer', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (3, 'Jane Smith', 'jane@example.com', '$2a$10$8K1p/a0gDt2FGnKaGzKGC.UMYx3KqK5DZ3FZh1Kk7U9f6h7.mQ2.m', 'customer', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    (4, 'Support Agent', 'agent@example.com', '$2a$10$8K1p/a0gDt2FGnKaGzKGC.UMYx3KqK5DZ3FZh1Kk7U9f6h7.mQ2.m', 'agent', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (email) DO NOTHING;

-- Reset sequence
SELECT setval('users_id_seq', COALESCE((SELECT MAX(id) FROM users), 1));
```

## Redis Configuration

### Redis Configuration File
```conf
# config/redis.conf
# Redis configuration for development

# Network
bind 127.0.0.1
port 6379
timeout 300

# General
daemonize no
pidfile /var/run/redis/redis-server.pid
loglevel notice
logfile ""

# Snapshotting
save 900 1
save 300 10
save 60 10000
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
dir /data

# Replication
replica-serve-stale-data yes
replica-read-only yes

# Memory Management
maxmemory 256mb
maxmemory-policy allkeys-lru

# Append Only File
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
no-appendfsync-on-rewrite no
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb

# Lua scripting
lua-time-limit 5000

# Slow log
slowlog-log-slower-than 10000
slowlog-max-len 128

# Event notification
notify-keyspace-events ""

# Advanced config
hash-max-ziplist-entries 512
hash-max-ziplist-value 64
list-max-ziplist-size -2
list-compress-depth 0
set-max-intset-entries 512
zset-max-ziplist-entries 128
zset-max-ziplist-value 64
hll-sparse-max-bytes 3000
stream-node-max-bytes 4096
stream-node-max-entries 100

# Client output buffer limits
client-output-buffer-limit normal 0 0 0
client-output-buffer-limit replica 256mb 64mb 60
client-output-buffer-limit pubsub 32mb 8mb 60

# Client query buffer limit
client-query-buffer-limit 1gb

# Protocol buffer limit
proto-max-bulk-len 512mb

# Frequency
hz 10

# Background task settings
dynamic-hz yes
aof-rewrite-incremental-fsync yes
rdb-save-incremental-fsync yes
```

## Monitoring Configuration

### Prometheus Configuration
```yaml
# config/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alerts.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['alertmanager:9093']

scrape_configs:
  - job_name: 'chat-widget-dashboard'
    static_configs:
      - targets: ['app:8080']
    metrics_path: '/metrics'
    scrape_interval: 10s
    scrape_timeout: 5s

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']

  - job_name: 'node'
    static_configs:
      - targets: ['node-exporter:9100']
```

### Grafana Dashboard Configuration
```yaml
# config/grafana/dashboards/dashboard.yml
apiVersion: 1

providers:
  - name: 'chat-widget-dashboards'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards
```

```yaml
# config/grafana/datasources/prometheus.yml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
```

## Development Scripts

### Database Management Script
```bash
#!/bin/bash
# scripts/db-manage.sh

set -euo pipefail

DATABASE_URL=${DATABASE_URL:-"postgres://chatuser:chatpass@localhost:5432/chatwidget?sslmode=disable"}
COMMAND=${1:-"help"}

case $COMMAND in
  "migrate")
    echo "Running database migrations..."
    go run ./cmd/server migrate
    ;;
  "rollback")
    echo "Rolling back last migration..."
    go run ./cmd/server migrate down
    ;;
  "seed")
    echo "Seeding database with test data..."
    psql $DATABASE_URL -f scripts/seed-data.sql
    ;;
  "reset")
    echo "Resetting database..."
    psql $DATABASE_URL -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
    go run ./cmd/server migrate
    psql $DATABASE_URL -f scripts/seed-data.sql
    ;;
  "backup")
    BACKUP_FILE="backup_$(date +%Y%m%d_%H%M%S).sql"
    echo "Creating backup: $BACKUP_FILE"
    pg_dump $DATABASE_URL > backups/$BACKUP_FILE
    ;;
  "restore")
    BACKUP_FILE=${2:-""}
    if [ -z "$BACKUP_FILE" ]; then
      echo "Usage: $0 restore <backup_file>"
      exit 1
    fi
    echo "Restoring from backup: $BACKUP_FILE"
    psql $DATABASE_URL < backups/$BACKUP_FILE
    ;;
  "help")
    echo "Database management script"
    echo "Usage: $0 <command>"
    echo ""
    echo "Commands:"
    echo "  migrate   - Run database migrations"
    echo "  rollback  - Rollback last migration"
    echo "  seed      - Seed database with test data"
    echo "  reset     - Reset database (drop all data and recreate)"
    echo "  backup    - Create database backup"
    echo "  restore   - Restore from backup file"
    echo "  help      - Show this help message"
    ;;
  *)
    echo "Unknown command: $COMMAND"
    echo "Run '$0 help' for usage information"
    exit 1
    ;;
esac
```

### Environment Setup Script
```bash
#!/bin/bash
# scripts/setup-dev.sh

set -euo pipefail

echo "Setting up development environment for Chat Widget Dashboard..."

# Check if required tools are installed
check_command() {
    if ! command -v $1 &> /dev/null; then
        echo "Error: $1 is not installed"
        exit 1
    fi
}

echo "Checking required tools..."
check_command "go"
check_command "docker"
check_command "docker-compose"
check_command "psql"

# Create necessary directories
echo "Creating project directories..."
mkdir -p uploads/models
mkdir -p logs
mkdir -p backups
mkdir -p tmp

# Copy environment file if it doesn't exist
if [ ! -f .env ]; then
    echo "Creating .env file from template..."
    cp .env.example .env
    echo "Please update .env file with your configuration"
fi

# Install Go dependencies
echo "Installing Go dependencies..."
go mod download
go mod tidy

# Install development tools
echo "Installing development tools..."
go install github.com/cosmtrek/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/go-delve/delve/cmd/dlv@latest

# Start database and redis
echo "Starting database and Redis..."
docker-compose up -d db redis

# Wait for database to be ready
echo "Waiting for database to be ready..."
while ! pg_isready -h localhost -p 5432 -U chatuser; do
    sleep 1
done

# Run database migrations
echo "Running database migrations..."
./scripts/db-manage.sh migrate

# Seed database with test data
echo "Seeding database with test data..."
./scripts/db-manage.sh seed

echo "✅ Development environment setup complete!"
echo ""
echo "Available commands:"
echo "  make dev          - Start development server with hot reload"
echo "  make test         - Run tests"
echo "  make lint         - Run linter"
echo "  make docker-run   - Start full development stack with Docker"
echo ""
echo "URLs:"
echo "  Application:      http://localhost:8080"
echo "  Database Admin:   http://localhost:8081"
echo "  Redis Admin:      http://localhost:8082"
echo "  Grafana:          http://localhost:3000 (admin/admin)"
echo "  Prometheus:       http://localhost:9090"
```

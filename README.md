# ⚡ QueueForge

A distributed job processing system built with Go. Submit a URL via the web UI or REST API — a worker pool fetches the page, counts words, and persists the result. Includes retry logic, graceful shutdown, Prometheus metrics, and Docker Compose for one-command deployment.

![Go](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat-square&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-blue?style=flat-square&logo=postgresql)
![Redis](https://img.shields.io/badge/Redis-7-red?style=flat-square&logo=redis)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat-square&logo=docker)
![Prometheus](https://img.shields.io/badge/Prometheus-Metrics-E6522C?style=flat-square&logo=prometheus)

---

## How It Works
POST /api/jobs
│
▼
PostgreSQL (QUEUED)
│
▼
Redis Queue (LPUSH)
│
▼
Worker Pool — 5 goroutines (BRPOP)
│
├── fetch URL
├── count words
└── PostgreSQL (COMPLETED / FAILED)

## Features

- **Concurrent worker pool** — configurable goroutine count via env var
- **Exponential backoff retry** — up to 3 attempts before marking FAILED
- **Graceful shutdown** — SIGTERM drains in-flight jobs cleanly
- **Prometheus metrics** — job count, duration histogram, queue size
- **Web UI** — live polling dashboard with pipeline step visualization
- **Docker Compose** — full stack in one command

## Quick Start

### Local (requires Go 1.23, PostgreSQL, Redis)

```bash
git clone https://github.com/ulasdursun/queueforge.git
cd queueforge
cp .env.example .env
go run ./cmd/server
```

Open `http://localhost:8080`

### Docker

```bash
docker-compose up --build
```

| Service    | URL                         |
|------------|-----------------------------|
| Web UI     | http://localhost:8080        |
| Prometheus | http://localhost:9090        |
| Grafana    | http://localhost:3000        |

Grafana default credentials: `admin / admin`

## API

| Method | Endpoint         | Description              |
|--------|-----------------|--------------------------|
| POST   | `/api/jobs`      | Create a URL fetch job   |
| GET    | `/api/jobs`      | List jobs (filter by status) |
| GET    | `/api/jobs/:id`  | Get job by ID            |
| GET    | `/health`        | Health check             |
| GET    | `/metrics`       | Prometheus metrics       |

### Create job

```bash
curl -X POST http://localhost:8080/api/jobs \
  -H "Content-Type: application/json" \
  -d '{"type":"URL_FETCH","payload":"{\"url\":\"https://nytimes.com\"}"}'
```

## Configuration

| Env var        | Default    | Description               |
|----------------|------------|---------------------------|
| SERVER_PORT    | 8080       | HTTP server port          |
| DATABASE_URL   | —          | PostgreSQL connection URL |
| REDIS_URL      | —          | Redis connection URL      |
| WORKER_COUNT   | 5          | Number of worker goroutines |
| MAX_RETRIES    | 3          | Max retry attempts per job |

## Project Structure
queueforge/
├── cmd/server/        — entry point, wiring
├── internal/
│   ├── api/           — Gin router, handlers, middleware
│   ├── domain/        — Job entity, status enum, errors
│   ├── repository/    — PostgreSQL queries
│   ├── queue/         — Redis push/pop
│   ├── worker/        — goroutine pool, processor, retry
│   └── metrics/       — Prometheus counters and histograms
├── config/            — env var loading
├── static/            — web UI
└── docker/            — init.sql, prometheus.yml

## Known Limitations

- Redis restart loses queued job IDs (jobs remain in DB as QUEUED)
- Single-node worker — no horizontal scaling
- Job type limited to URL_FETCH
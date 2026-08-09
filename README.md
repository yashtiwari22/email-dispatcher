# ⚡ Production High-Throughput Email Dispatcher Monorepo

> A production-grade, distributed email dispatching system built using a Go REST API Gateway, Redis-backed Asynq worker pools, GORM PostgreSQL storage, Server-Sent Events (SSE) real-time delivery tracking, and a Next.js 15 App Router web dashboard.

---

## 🏛 System Architecture

```text
               ┌─────────────────────────────────────────┐
               │    Next.js 15 App Router Web Dashboard   │
               │        (TailwindCSS + Lucide UI)         │
               └────────────────────┬────────────────────┘
                                    │
                                    │ HTTP REST API / SSE Stream
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│                   Go REST API Gateway (backend/api)                    │
│   ├── POST /api/v1/campaigns        ├── POST /api/v1/campaigns/upload │
│   ├── GET  /api/v1/campaigns/stream └── POST /api/v1/dlq/:id/replay    │
└──────────────────┬─────────────────────────────────┬───────────────────┘
                   │                                 │
     Enqueues Tasks│                                 │ GORM Queries
                   ▼                                 ▼
┌──────────────────────────────────┐      ┌──────────────────────────────┐
│  Redis 7 Asynq Task Queue        │      │ PostgreSQL 16 Storage        │
│  (Exponential Backoff & DLQ)     │      │ (Campaigns, Logs, DLQ)       │
└──────────────────┬───────────────┘      └──────────────┬───────────────┘
                   │                                     │
                   │ Dequeues Jobs                       │ Updates Status
                   ▼                                     │
┌────────────────────────────────────────────────────────┴───────────────┐
│              Asynq Worker Engine Pool (backend/engine)                │
│   ├── Thread-Safe Template Cache (sync.RWMutex + html/template)        │
│   └── Zero-Allocation SMTP Formatter (fmt.Appendf)                     │
└──────────────────┬─────────────────────────────────────────────────────┘
                   │
                   ▼
   ┌───────────────────────────────┐
   │    SMTP Server / Mailpit UI    │
   └───────────────────────────────┘
```

---

## ⚡ Key Engineering Highlights & Resume Bullet Points

- **High-Throughput Distributed Worker Queue**: Built on Redis-backed `Asynq`, delivering exponential backoff retries (3 retries max), timeout protection, and automatic Dead-Letter Queue (DLQ) capture for unrecoverable delivery errors.
- **Zero-Allocation SMTP Formatting**: Utilizes Go 1.19+ `fmt.Appendf(nil, ...)` to format RFC 822 email payloads directly into byte buffers, bypassing intermediate string allocations.
- **Thread-Safe Template Caching**: Implemented a `sync.RWMutex` concurrent compilation engine for Go `html/template`, ensuring zero template re-parsing overhead under heavy worker concurrency.
- **Chunked CSV Streaming Uploader**: Memory-efficient multipart CSV parser (`backend/api/handlers_csv.go`) processing large recipient lists in 250-row chunks to prevent heap spikes.
- **Real-Time Progress SSE Stream**: Built a Server-Sent Events (`GET /api/v1/campaigns/:id/stream`) endpoint pushing live progress metrics (`sent`, `failed`, `progress%`) to connected Next.js dashboard clients.
- **Interactive DLQ Management UI**: Visual inspector table allowing developers to inspect failed job JSON payloads and trigger 1-click re-enqueue replays (`POST /api/v1/dlq/:id/replay`).
- **GitHub Stacked PR Protocol**: Architected using GitHub's native stacked PR protocol (`gh stack`), partitioning the system into small, reviewable layers.

---

## 🥞 GitHub Stacked PR Architecture

```text
layer-6-deployment-cicd  → PR #7 (base: main)                        ← Current Layer
layer-5-nextjs-web       → PR #6 (base: main) [Merged]
layer-4-api-gateway      → PR #5 (base: main) [Merged]
layer-3-worker-engine     → PR #4 (base: main) [Merged]
layer-2-db-schema        → PR #2 (base: main) [Merged]
layer-1-infra-monorepo   → PR #1 (base: main) [Merged]
main (default trunk branch)
```

---

## 🛠 100% Free Local Quickstart

### Prerequisites
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [Go 1.22+](https://go.dev/dl/)
- [Node.js 20+ & pnpm](https://pnpm.io/)

### 1. Start Infrastructure (PostgreSQL, Redis, Mailpit)
```bash
docker-compose -f infra/docker-compose.yml up -d
```

### 2. Run All Go Backend Tests
```bash
go test -v ./backend/db ./backend/engine ./backend/api
```

### 3. Start Next.js 15 Web Dashboard
```bash
cd frontend
pnpm install
pnpm dev
```
Open [http://localhost:3000](http://localhost:3000) in your browser. Mailpit SMTP web UI is available at [http://localhost:8025](http://localhost:8025).

---

## 📁 Repository Structure

```text
.
├── .github/workflows/ci.yml     # Automated GitHub Actions CI workflow
├── backend/
│   ├── db/                      # GORM PostgreSQL entity models & SQLite test suite
│   ├── engine/                  # Asynq worker server, template cache & SMTP sender
│   └── api/                     # Go REST API gateway, CSV uploader & SSE stream
├── frontend/                    # Next.js 15 App Router web dashboard & DLQ UI
├── infra/
│   ├── docker-compose.yml       # Local dev services (Postgres 16, Redis 7, Mailpit)
│   ├── Dockerfile.api           # Production multi-stage API Dockerfile (<30MB)
│   ├── Dockerfile.web           # Production Next.js web Dockerfile
│   └── render.yaml              # Free-tier cloud container deployment spec
├── go.work                      # Native Go multi-module workspace definition
├── pnpm-workspace.yaml          # Monorepo pnpm workspace definition
└── TASKS.md                     # Master project task reference checklist
```

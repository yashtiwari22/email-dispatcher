# Task Checklist & Reference Guide: Production Email Dispatcher Monorepo

This document serves as the master task reference checklist for the step-by-step transformation of `email-dispatcher` into a production-grade full-stack monorepo built using **GitHub Stacked Pull Requests** (`gh stack`).

---

## 🔒 Security & Policy Invariant
> **CRITICAL**: The AI assistant MUST NEVER view, read, edit, or touch any actual `.env` files or secret credentials. All configuration uses `.env.example` templates with generic placeholders. Actual secrets are managed exclusively by the user.

---

## 🏁 Prerequisites: Manual Git Repository Setup

- [ ] **0.1 Initialize Git Repo**:
  ```bash
  git init -b main
  ```
- [ ] **0.2 Initial Baseline Commit**:
  ```bash
  git add .
  git commit -m "chore: initial commit (standalone Go baseline)"
  ```
- [ ] **0.3 Link Manually Created GitHub Remote & Push Main**:
  ```bash
  git remote add origin <YOUR_GITHUB_REPO_URL>
  git push -u origin main
  ```
- [ ] **0.4 Install gh-stack CLI Extension**:
  ```bash
  gh extension install github/gh-stack
  ```

---

## 🏛 Folder Architecture

- `backend/`
  - `backend/db/`: PostgreSQL database models, GORM connection manager, migrations & seed scripts.
  - `backend/engine/`: Redis Asynq task server, worker pool, thread-safe template engine & SMTP client.
  - `backend/api/`: Go REST API gateway, CSV streaming uploader, SSE progress stream, DLQ management.
- `frontend/`: Next.js 15 App Router web dashboard with real-time SSE progress bar, CSV upload modal & DLQ inspector.
- `infra/`: Docker Compose (Postgres 16, Redis 7, Mailpit) and cloud deployment configs.

---

## 🥞 GitHub Stacked PR Architecture & Core Rules

Official Reference: [GitHub Stacked PRs Documentation](https://docs.github.com/en/pull-requests/get-started/about-stacked-prs)

### Visual Dependency Chain
```text
layer-6-deployment-cicd  → PR #6 (base: layer-5-nextjs-web)       ← Top Layer
layer-5-nextjs-web       → PR #5 (base: layer-4-api-gateway)
layer-4-api-gateway      → PR #4 (base: layer-3-worker-engine)
layer-3-worker-engine     → PR #3 (base: layer-2-db-schema)
layer-2-db-schema        → PR #2 (base: layer-1-infra-monorepo)
layer-1-infra-monorepo   → PR #1 (base: main)                       ← Bottom Layer
main (default trunk branch)
```

---

## 🥞 Stack Layer 1: Monorepo Foundation & Dev Infrastructure
**Branch**: `layer-1-infra-monorepo` | **Base Branch Target**: `main` (`PR #1`)

- [ ] **1.1 Workspace Setup**: Create `pnpm-workspace.yaml` and `go.work` linking `backend/` modules and `frontend/`.
- [ ] **1.2 Build Pipeline**: Configure `turbo.json` with standard tasks (`build`, `dev`, `lint`, `test`).
- [ ] **1.3 Dev Container Services**: Create `infra/docker-compose.yml` orchestrating PostgreSQL 16, Redis 7, and Mailpit.
- [ ] **1.4 Environment Template**: Create `.env.example` with generic configuration placeholders (`DB_HOST`, `DB_PORT`, `REDIS_HOST`, `SMTP_PORT`).
- [ ] **1.5 Root Tooling**: Initialize root `package.json`, `.gitignore`, Prettier, and ESLint configs.
- [ ] **1.6 Layer 1 Stack Commit & Push**:
  ```bash
  git checkout -b layer-1-infra-monorepo main
  git add .
  git commit -m "feat(infra): setup monorepo foundation, docker compose & dev tooling"
  gh stack create --title "Layer 1: Monorepo Foundation & Dev Infrastructure" --body "Sets up Turborepo, Go workspace, and Docker Compose."
  ```

---

## 🥞 Stack Layer 2: Database Schema & Domain Data Layer (`backend/db`)
**Branch**: `layer-2-db-schema` | **Base Branch Target**: `layer-1-infra-monorepo` (`PR #2`)

- [ ] **2.1 Package Structure**: Initialize `backend/db/` with Go module structure.
- [ ] **2.2 Schema Definitions**: Define PostgreSQL tables & models for `campaigns`, `recipients`, `email_templates`, `dispatch_logs`, `dlq_records`.
- [ ] **2.3 Data Layer & Migrations**: Implement GORM connection manager, auto-migrations, and connection pooling.
- [ ] **2.4 Seeding Utility**: Create database seed script (`backend/db/seed.go`) with realistic mock datasets.
- [ ] **2.5 Automated Tests**: Add unit tests for DB initialization, CRUD operations, and transaction support (`backend/db/db_test.go`).
- [ ] **2.6 Layer 2 Stack Commit & Push**:
  ```bash
  git checkout -b layer-2-db-schema layer-1-infra-monorepo
  git add .
  git commit -m "feat(db): implement GORM PostgreSQL schema, migrations and seeding"
  gh stack create --title "Layer 2: Database Schema & Domain Data Layer" --parent layer-1-infra-monorepo
  ```

---

## 🥞 Stack Layer 3: High-Throughput Dispatcher Engine & Asynq Worker Pool (`backend/engine`)
**Branch**: `layer-3-worker-engine` | **Base Branch Target**: `layer-2-db-schema` (`PR #3`)

- [ ] **3.1 Worker Server**: Implement `backend/engine/` using Redis-backed `Asynq` queue with concurrency limits and rate limiters.
- [ ] **3.2 Template Caching**: Implement thread-safe `html/template` caching engine (parse once at startup, reuse in-memory).
- [ ] **3.3 SMTP Connection Pool**: Implement reusable SMTP connection pool (`net/smtp` + TLS/auth support for Mailpit and production providers).
- [ ] **3.4 Resilience & DLQ**: Implement exponential backoff retry policy and automatic dead-letter queue (DLQ) capture for failed jobs.
- [ ] **3.5 Engine Tests**: Add unit & integration tests for job dispatching, template execution, SMTP client, and DLQ capture (`backend/engine/template_test.go`).
- [ ] **3.6 Layer 3 Stack Commit & Push**:
  ```bash
  git checkout -b layer-3-worker-engine layer-2-db-schema
  git add .
  git commit -m "feat(engine): implement Asynq Redis worker pool, template engine and DLQ"
  gh stack create --title "Layer 3: High-Throughput Worker Engine & Asynq Queue" --parent layer-2-db-schema
  ```

---

## 🥞 Stack Layer 4: REST API Gateway & Real-Time SSE Stream (`backend/api`)
**Branch**: `layer-4-api-gateway` | **Base Branch Target**: `layer-3-worker-engine` (`PR #4`)

- [ ] **4.1 API Framework**: Initialize `backend/api/` Go web service with Gin framework and CORS middleware.
- [ ] **4.2 CSV Streaming Uploader**: Implement `POST /api/v1/campaigns/upload` for chunked CSV parsing and instant batch job queuing.
- [ ] **4.3 Campaign Endpoints**: Implement Campaign CRUD (`POST`, `GET`, `GET /:id`, `PATCH /:id/status`).
- [ ] **4.4 Real-Time Progress SSE**: Implement Server-Sent Events (`GET /api/v1/campaigns/:id/stream`) pushing live progress metrics to clients.
- [ ] **4.5 DLQ Replay API**: Implement DLQ endpoints (`GET /api/v1/dlq`, `POST /api/v1/dlq/:id/replay`, `DELETE /api/v1/dlq`).
- [ ] **4.6 Observability**: Add `/healthz` and `/readyz` probes with structured JSON logging.
- [ ] **4.7 API Tests**: Write HTTP integration tests for endpoints and stream handlers (`backend/api/handlers_test.go`).
- [ ] **4.8 Layer 4 Stack Commit & Push**:
  ```bash
  git checkout -b layer-4-api-gateway layer-3-worker-engine
  git add .
  git commit -m "feat(api): implement Go REST API gateway, CSV uploader and SSE stream"
  gh stack create --title "Layer 4: REST API Gateway & Real-Time SSE Stream" --parent layer-3-worker-engine
  ```

---

## 🥞 Stack Layer 5: Modern Next.js 15 Dashboard (`frontend`)
**Branch**: `layer-5-nextjs-web` | **Base Branch Target**: `layer-4-api-gateway` (`PR #5`)

- [ ] **5.1 App Setup**: Initialize `frontend/` Next.js 15 (App Router, React 19, TypeScript, Tailwind CSS v4, Shadcn UI).
- [ ] **5.2 Dashboard Overview**: Build real-time analytics dashboard with SSE live progress bars, throughput metrics, and campaign status badges.
- [ ] **5.3 CSV Drag-and-Drop Uploader**: Build interactive CSV uploader modal with column auto-mapping and preview table.
- [ ] **5.4 DLQ Management UI**: Build visual DLQ inspector table with error detail modal and bulk retry triggers.
- [ ] **5.5 Template Editor**: Build rich template builder with live email preview and dynamic variable tags (`{{.Name}}`, `{{.Email}}`).
- [ ] **5.6 Frontend Tests**: Add component and integration tests using Vitest/React Testing Library.
- [ ] **5.7 Layer 5 Stack Commit & Push**:
  ```bash
  git checkout -b layer-5-nextjs-web layer-4-api-gateway
  git add .
  git commit -m "feat(web): implement Next.js 15 App Router dashboard & DLQ UI"
  gh stack create --title "Layer 5: Modern Next.js 15 Dashboard" --parent layer-4-api-gateway
  ```

---

## 🥞 Stack Layer 6: CI/CD Pipeline & Cloud Deployment Setup (`infra/` & `.github/`)
**Branch**: `layer-6-deployment-cicd` | **Base Branch Target**: `layer-5-nextjs-web` (`PR #6`)

- [ ] **6.1 GitHub Actions**: Create `.github/workflows/ci.yml` running linting, Go tests, Next.js build, and `gh stack` checks.
- [ ] **6.2 Multi-Stage Dockerfiles**: Create slim production Dockerfiles for `backend/api` (`infra/Dockerfile.api`) and `frontend` (`infra/Dockerfile.web`).
- [ ] **6.3 Cloud Deployment Configuration**: Add deployment blueprint (`infra/render.yaml`) for Render / Railway / Cloud containers.
- [ ] **6.4 Showcase Documentation**: Write top-tier `README.md` with visual architecture diagrams, resume key highlights, local quickstart guide, and live demo links.
- [ ] **6.5 Layer 6 Stack Commit & Push**:
  ```bash
  git checkout -b layer-6-deployment-cicd layer-5-nextjs-web
  git add .
  git commit -m "ci(deploy): add GitHub Actions pipeline, Dockerfiles & cloud deployment blueprint"
  gh stack create --title "Layer 6: CI/CD Pipeline & Cloud Deployment Setup" --parent layer-5-nextjs-web
  gh stack submit
  ```

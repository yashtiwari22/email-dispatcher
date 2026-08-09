# Task Checklist & Reference Guide: Production Email Dispatcher Monorepo

This document serves as the master task reference checklist for the step-by-step transformation of `email-dispatcher` into a production-grade full-stack monorepo built using **GitHub Stacked Pull Requests** (`gh stack`) and native CLI commands.

---

## 🔒 Security & Policy Invariants
> **CRITICAL SECRETS INVARIANT**: The AI assistant MUST NEVER view, read, edit, or touch any actual `.env` files or secret credentials. All configuration uses `.env.example` templates with generic placeholders. Actual secrets are managed exclusively by the user.

> **PREFER NATIVE CLI TOOLING**: Use standard CLI commands (`go mod init`, `go work init`, `go work use`, `pnpm create next-app`) for module/package initialization rather than manually writing empty stub files.

---

## 🛠 Official GitHub Stacked PR CLI Commands (`gh stack`)

From [GitHub Official Stacked PRs Quickstart](https://docs.github.com/en/pull-requests/get-started/stacked-prs-quickstart):

- `gh stack init`: Initialize stack tracking on default branch (`main`).
- `gh stack add BRANCH-NAME`: Add a new dependent branch to the top of the stack.
- `gh stack push`: Push all stack branches to GitHub.
- `gh stack submit`: Create and link pull requests on GitHub with automatic base branch targeting.
- `gh stack view`: View visual tree of all branches, PR links, and statuses.

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

- [x] **1.1 Workspace Setup**: Run `go work init .` and create `pnpm-workspace.yaml` linking `backend/` modules and `frontend/`.
- [x] **1.2 Build Pipeline**: Configure `turbo.json` with standard tasks (`build`, `dev`, `lint`, `test`).
- [x] **1.3 Dev Container Services**: Create `infra/docker-compose.yml` orchestrating PostgreSQL 16, Redis 7, and Mailpit.
- [x] **1.4 Environment Template**: Create `.env.example` with generic configuration placeholders (`DB_HOST`, `DB_PORT`, `REDIS_HOST`, `SMTP_PORT`).
- [x] **1.5 Root Tooling**: Initialize root `package.json`, `.gitignore`, Prettier, and ESLint configs.
- [x] **1.6 Layer 1 Stack Commit & Push**:
  ```bash
  gh stack add layer-1-infra-monorepo
  go work init .
  git add .
  git commit -m "feat(infra): setup monorepo foundation, docker compose & dev tooling"
  ```

---

## 🥞 Stack Layer 2: Database Schema & Domain Data Layer (`backend/db`)
**Branch**: `layer-2-db-schema` | **Base Branch Target**: `layer-1-infra-monorepo` (`PR #2`)

- [x] **2.1 Module Initialization**: Initialize `backend/db` natively via `go mod init` and `go work use ./backend/db`.
- [x] **2.2 Schema Definitions**: Define PostgreSQL tables & models for `campaigns`, `recipients`, `email_templates`, `dispatch_logs`, `dlq_records`.
- [x] **2.3 Data Layer & Migrations**: Implement GORM connection manager, auto-migrations, and connection pooling.
- [x] **2.4 Seeding Utility**: Create database seed script (`backend/db/seed.go`) with realistic mock datasets.
- [x] **2.5 Automated Tests**: Add unit tests for DB initialization, CRUD operations, and transaction support (`backend/db/db_test.go`).
- [ ] **2.6 Layer 2 Stack Commit & Push**:
  ```bash
  gh stack add layer-2-db-schema
  mkdir -p backend/db && cd backend/db && go mod init github.com/yashtiwari22/email-dispatcher/backend/db
  go work use ./backend/db
  git add .
  git commit -m "feat(db): implement GORM PostgreSQL schema, migrations and seeding"
  ```

---

## 🥞 Stack Layer 3: High-Throughput Dispatcher Engine & Asynq Worker Pool (`backend/engine`)
**Branch**: `layer-3-worker-engine` | **Base Branch Target**: `layer-2-db-schema` (`PR #3`)

- [x] **3.1 Module Initialization**: Initialize `backend/engine` natively via `go mod init` and `go work use ./backend/engine`.
- [x] **3.2 Worker Server**: Implement `backend/engine/` using Redis-backed `Asynq` queue with concurrency limits and rate limiters.
- [x] **3.3 Template Caching & SMTP**: Implement thread-safe `html/template` caching engine and SMTP connection pool.
- [x] **3.4 Resilience & DLQ**: Implement exponential backoff retry policy and automatic dead-letter queue (DLQ) capture for failed jobs.
- [x] **3.5 Engine Tests**: Add unit & integration tests for job dispatching, template execution, SMTP client, and DLQ capture (`backend/engine/template_test.go`).
- [x] **3.6 Layer 3 Stack Commit & Push**:
  ```bash
  gh stack add layer-3-worker-engine
  mkdir -p backend/engine && cd backend/engine && go mod init github.com/yashtiwari22/email-dispatcher/backend/engine
  go work use ./backend/engine
  git add .
  git commit -m "feat(engine): implement Asynq Redis worker pool, template engine and DLQ"
  ```

---

## 🥞 Stack Layer 4: REST API Gateway & Real-Time SSE Stream (`backend/api`)
**Branch**: `layer-4-api-gateway` | **Base Branch Target**: `layer-3-worker-engine` (`PR #4`)

- [x] **4.1 Module Initialization**: Initialize `backend/api` natively via `go mod init` and `go work use ./backend/api`.
- [x] **4.2 CSV Streaming Uploader**: Implement `POST /api/v1/campaigns/upload` for chunked CSV parsing and instant batch job queuing.
- [x] **4.3 Campaign Endpoints**: Implement Campaign CRUD (`POST`, `GET`, `GET /:id`, `PATCH /:id/status`).
- [x] **4.4 Real-Time Progress SSE**: Implement Server-Sent Events (`GET /api/v1/campaigns/:id/stream`) pushing live progress metrics to clients.
- [x] **4.5 DLQ Replay API**: Implement DLQ endpoints (`GET /api/v1/dlq`, `POST /api/v1/dlq/:id/replay`, `DELETE /api/v1/dlq`).
- [x] **4.6 Observability**: Add `/healthz` and `/readyz` probes with structured JSON logging.
- [x] **4.7 API Tests**: Write HTTP integration tests for endpoints and stream handlers (`backend/api/handlers_test.go`).
- [ ] **4.8 Layer 4 Stack Commit & Push**:
  ```bash
  gh stack add layer-4-api-gateway
  mkdir -p backend/api && cd backend/api && go mod init github.com/yashtiwari22/email-dispatcher/backend/api
  go work use ./backend/api
  git add .
  git commit -m "feat(api): implement Go REST API gateway, CSV uploader and SSE stream"
  ```

---

## 🥞 Stack Layer 5: Modern Next.js 15 Dashboard (`frontend`)
**Branch**: `layer-5-nextjs-web` | **Base Branch Target**: `layer-4-api-gateway` (`PR #5`)

- [x] **5.1 App Setup**: Initialize `frontend/` natively via `pnpm create next-app frontend --typescript --tailwind --app --eslint`.
- [x] **5.2 Dashboard Overview**: Build real-time analytics dashboard with SSE live progress bars, throughput metrics, and campaign status badges.
- [x] **5.3 CSV Drag-and-Drop Uploader**: Build interactive CSV uploader modal with column auto-mapping and preview table.
- [x] **5.4 DLQ Management UI**: Build visual DLQ inspector table with error detail modal and bulk retry triggers.
- [x] **5.5 Template Editor**: Build rich template builder with live email preview and dynamic variable tags (`{{.Name}}`, `{{.Email}}`).
- [x] **5.6 Frontend Build Verification**: Verify TypeScript types and production bundle compilation (`pnpm build`).
- [ ] **5.7 Layer 5 Stack Commit & Push**:
  ```bash
  gh stack add layer-5-nextjs-web
  pnpm create next-app frontend --typescript --tailwind --app --eslint
  git add .
  git commit -m "feat(web): implement Next.js 15 App Router dashboard & DLQ UI"
  ```

---

## 🥞 Stack Layer 6: CI/CD Pipeline & Cloud Deployment Setup (`infra/` & `.github/`)
**Branch**: `layer-6-deployment-cicd` | **Base Branch Target**: `layer-5-nextjs-web` (`PR #6`)

- [x] **6.1 GitHub Actions**: Create `.github/workflows/ci.yml` running linting, Go tests, Next.js build, and `gh stack` checks.
- [x] **6.2 Multi-Stage Dockerfiles**: Create slim production Dockerfiles for `backend/api` (`infra/Dockerfile.api`) and `frontend` (`infra/Dockerfile.web`).
- [x] **6.3 Cloud Deployment Configuration**: Add deployment blueprint (`infra/render.yaml`) for Render / Railway / Cloud containers.
- [x] **6.4 Showcase Documentation**: Write top-tier `README.md` with visual architecture diagrams, resume key highlights, local quickstart guide, and live demo links.
- [x] **6.5 Layer 6 Stack Commit & Push**:
  ```bash
  gh stack add layer-6-deployment-cicd
  git add .
  git commit -m "ci(deploy): add GitHub Actions pipeline, Dockerfiles & cloud deployment blueprint"
  gh stack push
  gh stack submit
  ```

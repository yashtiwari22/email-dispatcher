# Workspace Rules for Email Dispatcher Monorepo

## 🛑 STRICT EXPLICIT APPROVAL FOR FILE GENERATION

1. **NO AUTO-EXECUTION AFTER TERMINAL STEPS**:
   - The AI assistant MUST **NEVER** invoke code modification or file creation tools (`write_to_file`, `replace_file_content`, etc.) automatically after a terminal command, branch switch, or status update.
   - The assistant MUST pause and wait for an explicit user prompt (e.g., "create the files", "generate Layer 1", "proceed") before writing, editing, or creating any files.

---

## ⚡ PREFER NATIVE CLI COMMANDS & IDIOMATIC GENERATION OVER MANUAL STUBS

2. **USE NATIVE TOOLING FOR PACKAGE & MODULE INITIALIZATION**:
   - Do NOT manually write empty stub files (`go.mod`, `go.work`, `package.json`) when native CLI commands exist.
   - Always use standard idiomatic CLI commands for initializing modules and workspaces:
     - Go Modules: `go mod init <module-path>`
     - Go Workspace: `go work init` / `go work use <path>`
     - Node/pnpm: `pnpm init`, `pnpm create next-app`
   - The assistant should provide copy-pasteable native terminal commands or run them when approved.

---

## 🚨 CRITICAL SECURITY & SECRETS INVARIANT

3. **NEVER TOUCH OR VIEW SECRETS / ENV FILES**:
   - The AI assistant must **NEVER** view, read, grep, edit, create, or modify any `.env`, `.env.local`, `.env.production`, `.env.development`, secrets files, private keys, credentials, or token files containing actual values.
   - The AI assistant may ONLY create and update `.env.example` template files containing generic, non-sensitive placeholder values (e.g., `DB_HOST=localhost`, `DB_PORT=5432`, `SMTP_PORT=1025`).
   - Creation and management of actual `.env` files and environment secrets is strictly reserved for the human developer.

---

## 🏛 Monorepo & Architecture Invariants

4. **Strict Monorepo & Architectural Boundaries**:
   - Every service (`frontend/`, `backend/db`, `backend/engine`, `backend/api`, `infra/`) must reside in dedicated directories within the monorepo.
   - Do not write monolithic inline scripts; separate domain logic, handlers, data access, and worker pools cleanly.

5. **GitHub Stacked PR Protocol**:
   - Utilize GitHub's native stacked PR workflow (`gh stack` CLI extension).
   - Break every feature into small, atomic PR layers as defined in `TASKS.md` and `implementation_plan.md`.

6. **Production Standard Invariants**:
   - Implement production-grade queueing (Redis-backed Asynq) with automatic retry mechanisms, exponential backoff, and a Dead Letter Queue (DLQ).
   - Implement structured JSON logging, health probes (`/healthz`, `/readyz`).

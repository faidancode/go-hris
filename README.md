# go-hris

Production-oriented HRIS backend API built with Go, Gin, PostgreSQL, Redis, and Casbin RBAC.

This repository demonstrates modular API design, transaction-safe business logic, role-based authorization, idempotent write flows, and practical test strategy for a multi-tenant HR system.

## Highlights

- Modular domain packages: `auth`, `employee`, `department`, `position`, `employee-salary`, `leave`, `payroll`, `rbac`
- Multi-tenant guardrails via `company_id` scoping in service/repository layer
- RBAC authorization with Casbin policies loaded per company
- Idempotency support for critical write endpoint (`POST /payrolls`) using Redis lock + response cache
- Audit-aware workflow fields (`created_by`, `approved_by`, `approved_at`) for approval-driven domains
- Explicit SQL migrations with FK behavior and indexing strategy
- Unit tests for service and handler layers with success/failure scenarios

## Architecture

The codebase follows a layered modular pattern:

1. `routes` layer: HTTP route registration + middleware composition
2. `handler` layer: request binding, context extraction, response handling
3. `service` layer: business rules, validation, transaction orchestration
4. `repo` layer: data access and query responsibility
5. `entity/dto`: domain model and API contract separation

Dependency wiring is centralized in `internal/app/registry.go`, and infrastructure bootstrap lives in `internal/app/app.go`.

## Tech Stack and Why

- `github.com/gin-gonic/gin`
  - Fast, minimal HTTP framework with clean middleware model.
- `gorm.io/gorm` + `gorm.io/driver/postgres`
  - Productive ORM with explicit query control, soft delete support, and PostgreSQL integration.
- `github.com/redis/go-redis/v9`
  - Reliable Redis client used for idempotency lock/caching and scalable write protection.
- `github.com/casbin/casbin/v2`
  - Policy-based authorization for flexible RBAC and multi-tenant access rules.
- `github.com/golang-jwt/jwt/v5`
  - JWT-based authentication with access/refresh token flow.
- `golang.org/x/crypto`
  - Password hashing (`bcrypt`) for secure credential handling.
- `github.com/google/uuid`
  - UUID handling across entities and API payload validation.
- `github.com/joho/godotenv`
  - Local env bootstrap for developer workflow consistency.
- `github.com/DATA-DOG/go-sqlmock`
  - Deterministic transaction-level unit testing for service layer.
- `go.uber.org/mock` + `github.com/stretchr/testify`
  - Mock generation and expressive assertions for robust unit tests.

## API Surface (Current)

Base path: `/api/v1`

- `auth`: login, refresh, register, me, logout
- `department`: CRUD
- `position`: CRUD
- `employee`: read/list/create
- `employee-salaries`: CRUD
- `leave`: CRUD + approval workflow fields
- `payroll`: CRUD + idempotent create
- `rbac`: enforce endpoint (`/rbac/enforce`)

A ready-to-import Postman collection is available at:

- `postman/go-hris.postman_collection.json`

## Data & Migration Strategy

Migrations are maintained in:

- `internal/shared/database/migrations`

Design choices include:

- explicit foreign key behavior (`CASCADE`, `RESTRICT`, `SET NULL` where appropriate)
- domain-specific indexes for read/write patterns (company/status, employee/date, soft-delete)
- workflow tables include audit ownership and approval metadata

## Security & Reliability Notes

- Authentication uses JWT access + refresh strategy.
- Authorization uses middleware RBAC checks per resource/action.
- Sensitive write flows can enforce idempotency to prevent duplicate processing.
- Service layer uses SQL transactions for consistency in state transitions.
- Validation and parsing are handled before persistence to protect data integrity.

## Project Structure

```text
cmd/api                # application entrypoint
internal/app           # bootstrap + module registry
internal/<module>      # per-domain package (entity/dto/repo/service/handler/routes)
internal/middleware    # auth, rbac, idempotency, request guards
internal/shared        # connection, common utilities, migrations
postman/               # API collection for manual testing
```

## Run Locally

### 1) Prerequisites

- Go (matching `go.mod`)
- PostgreSQL
- Redis

### 2) Configure environment

Set `.env` (minimum):

```env
APP_ENV=development
PORT=3000
DB_HOST=localhost
DB_USER=...
DB_PASSWORD=...
DB_NAME=...
DB_PORT=5432
DB_SSLMODE=disable
JWT_SECRET=...
REDIS_ADDR=localhost:6379
```

### 3) Start infra (optional via Docker)

```bash
make docker-infra
```

### 4) Run migrations

```bash
make migrate-up
```

### 5) Start API

```bash
make run
```

## Testing

```bash
go test ./...
```

Tests emphasize:

- service-layer business rules and transactional behavior
- handler-layer request/response behavior and validation paths
- positive and negative scenarios for critical flows

## Engineering Positioning

This codebase is intentionally structured to show practical mid-to-senior backend competencies:

- clear modular boundaries and dependency wiring
- explicit trade-off decisions around consistency, authorization, and auditability
- predictable error handling and testability
- production-minded concerns: migration hygiene, retry strategy, and failure surfaces

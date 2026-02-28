# Outbox Pattern — Production-Grade Distributed Consistency in Go

> A study of how high-scale financial systems guarantee **zero event loss** under failure, using the Transactional Outbox Pattern implemented with Go, PostgreSQL, and RabbitMQ.

---

## Table of Contents

- [Overview](#overview)
- [The Problem: Dual-Write](#the-problem-dual-write)
- [The Solution: Transactional Outbox](#the-solution-transactional-outbox)
- [Architecture](#architecture)
- [Services](#services)
  - [Transaction Service](#transaction-service)
  - [Users Service](#users-service)
- [API Reference](#api-reference)
- [Database Schema](#database-schema)
- [Resilience & Reliability](#resilience--reliability)
- [Observability](#observability)
- [CI/CD Pipeline](#cicd-pipeline)
- [Load Testing](#load-testing)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Project Structure](#project-structure)
- [Design Decisions](#design-decisions)

---

## Overview

This monorepo implements the **Transactional Outbox Pattern** across two independent Go microservices. The pattern solves the fundamental distributed systems problem of keeping a database write and a message broker publish **atomically consistent** — without distributed transactions.

**What you'll find here:**

| Feature | Detail |
|---------|--------|
| Language | Go 1.22+ |
| Architecture | Clean / Hexagonal (ports & adapters) |
| Database | PostgreSQL 16 |
| Broker | RabbitMQ 3.12 |
| Observability | Prometheus · Grafana · Loki · Promtail |
| Testing | Unit · Handler · Race-detector · 90% coverage gate |
| Load testing | k6 — p90 < 30 ms, error rate < 1% |
| CI | GitHub Actions (lint · format · test · build) |
| Code review | Claude Code automated PR reviews |

---

## The Problem: Dual-Write

In an event-driven system, a service must do two things after a business operation:

1. **Persist** the state change to the database
2. **Publish** a domain event to the message broker

These are two separate I/O operations. If the broker is down, overloaded, or the process crashes between steps 1 and 2, the event is **silently lost** — the database is updated but no consumer is notified.

```
┌─────────────────────────────────────────────────────────────┐
│                     THE DUAL-WRITE PROBLEM                   │
│                                                             │
│  Service  ──[1]──▶  Database    ✓  committed                │
│     │                                                       │
│     └──────[2]──▶  RabbitMQ    ✗  broker unavailable       │
│                                                             │
│  Result: database updated, event lost. System inconsistent. │
└─────────────────────────────────────────────────────────────┘
```

---

## The Solution: Transactional Outbox

Instead of publishing directly to the broker, the event is written to an **outbox table inside the same database transaction** as the business operation. A separate background worker then reads from the outbox and publishes to the broker.

```
┌─────────────────────────────────────────────────────────────────────┐
│                    TRANSACTIONAL OUTBOX FLOW                        │
│                                                                     │
│  ┌──────────┐   BEGIN TRANSACTION                                   │
│  │  Service │──▶ INSERT INTO transactions (...)  ─┐ atomic          │
│  └──────────┘   INSERT INTO outbox (...)          ─┘ commit         │
│                                                                     │
│  ┌──────────┐   SELECT * FROM outbox WHERE status='PENDING'         │
│  │  Worker  │──▶ FOR UPDATE SKIP LOCKED  (no concurrent races)      │
│  └────┬─────┘   UPDATE outbox SET status='PROCESSING'               │
│       │                                                             │
│       ├──▶ Publish to RabbitMQ  ✓  ──▶ UPDATE status='PROCESSED'   │
│       │                                                             │
│       └──▶ Publish fails        ✗  ──▶ MarkForRetry()              │
│                                         retry_count + 1             │
│                                         PENDING  (up to 3×)         │
│                                         FAILED   (after 3×)         │
└─────────────────────────────────────────────────────────────────────┘
```

**Guarantees:**
- ✅ Atomic persistence — domain data + event in one transaction
- ✅ No event loss — if the process crashes, the worker retries on restart
- ✅ At-least-once delivery — consumers must be idempotent
- ✅ Retry with exhaustion — up to 3 attempts, then marked FAILED
- ✅ Concurrent-safe — `FOR UPDATE SKIP LOCKED` prevents double-processing

---

## Architecture

Each service follows **Clean / Hexagonal Architecture**. Dependencies always point inward: the domain knows nothing about infrastructure.

```
┌─────────────────────────────────────────────────────────┐
│                     HEXAGONAL ARCHITECTURE               │
│                                                         │
│   HTTP Request                                          │
│        │                                                │
│        ▼                                                │
│  ┌──────────────┐                                       │
│  │   Handler    │  ← presentation layer                 │
│  └──────┬───────┘                                       │
│         │                                               │
│         ▼                                               │
│  ┌──────────────┐                                       │
│  │   Use Case   │  ← application / business logic       │
│  └──────┬───────┘                                       │
│         │  (depends only on interfaces)                 │
│         ▼                                               │
│  ┌──────────────┐     ┌──────────────┐                  │
│  │    Ports     │     │    Ports     │                  │
│  │ (interfaces) │     │ (interfaces) │                  │
│  └──────┬───────┘     └──────┬───────┘                  │
│         │                   │                           │
│         ▼                   ▼                           │
│  ┌──────────────┐     ┌──────────────┐                  │
│  │  Repository  │     │  Publisher   │                  │
│  │  (Postgres)  │     │  (RabbitMQ)  │                  │
│  └──────────────┘     └──────────────┘                  │
│      infra / adapters                                   │
└─────────────────────────────────────────────────────────┘
```

**Dependency rule:** `handler → usecase → ports ← infra`
The domain and use case layers have **zero imports** from infrastructure packages.

---

## Services

### Transaction Service

Manages financial transfers between users with full outbox event publishing.

**Key capabilities:**
- Create transactions atomically with outbox event in one DB transaction
- Idempotent `POST` via `Idempotency-Key` header (Stripe-style)
- Background worker with retry logic (up to 3 attempts before `FAILED`)
- RabbitMQ publisher using `event.ID` as `MessageId` for consumer-side deduplication
- Prometheus metrics exposed at `/metrics`
- Swagger UI at `/swagger/`

**Components:**

```
transaction-service/
├── cmd/
│   ├── main.go               # HTTP server wiring + graceful shutdown
│   └── worker/main.go        # Standalone outbox worker process
├── config/                   # Env-based configuration
├── infra/
│   ├── db/                   # PostgreSQL connection + pool tuning
│   └── repository/           # Postgres implementations
│       ├── transaction_repository.go
│       └── outbox_repository.go
├── internal/core/
│   ├── broker/               # RabbitMQ connection + publisher
│   ├── domain/
│   │   ├── entity/           # Transaction, Outbox — pure domain types
│   │   └── ports/            # Repository and publisher interfaces
│   ├── errors/               # Typed *Exception error pattern
│   ├── handler/              # HTTP handlers + factory
│   └── usecase/              # Business logic + factory
│       ├── create_transaction.go
│       ├── get_transaction_status.go
│       ├── get_balance.go
│       └── factory.go        # Single-call dependency wiring
├── migrations/               # SQL migration files
├── observability/            # Prometheus, Grafana, Loki config
└── tests/load/               # k6 load test scripts
```

### Users Service

Manages user registration with its own outbox pattern implementation.

**Key capabilities:**
- Create users with email validation and password hashing
- Atomic user + outbox event write
- Configurable background worker interval

---

## API Reference

### Transaction Service  `localhost:8080`

#### Create Transaction

```http
POST /api/v1/transactions
Content-Type: application/json
Idempotency-Key: <uuid>          (optional — enables idempotent requests)
```

```json
{
  "from_user_id": "user-abc",
  "to_user_id":   "user-xyz",
  "amount":       1000,
  "description":  "payment for services"
}
```

| Response | Condition |
|----------|-----------|
| `201 Created` | New transaction created |
| `200 OK` | Duplicate request with same `Idempotency-Key` |
| `400 Bad Request` | Validation failure (same user, invalid amount, missing fields) |
| `500 Internal Server Error` | Unexpected persistence error |

```json
{
  "code": 201,
  "message": "transaction created",
  "data": {
    "id":     "8d3a1f2c-...",
    "status": "PENDING"
  }
}
```

**Idempotency behaviour:**
If you send the same `Idempotency-Key` twice, the second request returns the **original transaction** with `200 OK` — no duplicate is created, no outbox event is emitted.

```bash
# First call — creates the transaction
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: my-key-001" \
  -d '{"from_user_id":"user-a","to_user_id":"user-b","amount":500}'
# → 201 Created

# Second call with same key — idempotent
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: my-key-001" \
  -d '{"from_user_id":"user-a","to_user_id":"user-b","amount":500}'
# → 200 OK  (same transaction ID returned)
```

---

#### Get Transaction Status

```http
GET /api/v1/transactions/{id}
```

```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "id":     "8d3a1f2c-...",
    "status": "PENDING"
  }
}
```

Transaction lifecycle: `PENDING` → `PROCESSING` → `PROCESSED` | `FAILED`

---

#### Get User Balance

```http
GET /api/v1/balance/{userId}
```

Returns the net balance for a user — sum of all `COMPLETED` transactions received minus all sent.

```json
{
  "code": 200,
  "message": "ok",
  "data": {
    "user_id": "user-abc",
    "balance": 4500
  }
}
```

---

#### Utility Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /metrics` | Prometheus metrics |
| `GET /swagger/` | Swagger UI |
| `GET /swagger/doc.json` | OpenAPI spec |

---

## Database Schema

### `transactions`

```sql
CREATE TABLE transactions (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    amount             BIGINT      NOT NULL CHECK (amount >= 0),
    description        TEXT        NOT NULL,
    from_user_id       UUID        NOT NULL,
    to_user_id         UUID        NOT NULL,
    transaction_status VARCHAR(50) NOT NULL,
    idempotency_key    VARCHAR(255) NULL,       -- unique partial index
    created_at         TIMESTAMP   NOT NULL DEFAULT NOW(),
    processed_at       TIMESTAMP   NULL,

    CONSTRAINT chk_transactions_users_different
        CHECK (from_user_id <> to_user_id)
);

-- Indexes for read performance
CREATE INDEX idx_transactions_from_user_id   ON transactions (from_user_id);
CREATE INDEX idx_transactions_to_user_id     ON transactions (to_user_id);
CREATE INDEX idx_transactions_status         ON transactions (transaction_status);
CREATE INDEX idx_transactions_created_at     ON transactions (created_at);

-- Partial unique index — enforces idempotency at DB level
CREATE UNIQUE INDEX uidx_transactions_idempotency_key
    ON transactions (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
```

### `outbox`

```sql
CREATE TABLE outbox (
    id           UUID        PRIMARY KEY,
    type         VARCHAR(200) NOT NULL,   -- e.g. "TransactionCreated"
    payload      TEXT        NOT NULL,   -- JSON event body
    status       VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    retry_count  INT         NOT NULL DEFAULT 0,
    created_at   TIMESTAMP   NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP   NULL
);

CREATE INDEX idx_outbox_status     ON outbox (status);
CREATE INDEX idx_outbox_created_at ON outbox (created_at);
```

**Outbox status transitions:**

```
PENDING ──▶ PROCESSING ──▶ PROCESSED   (success)
                  │
                  └──▶ PENDING         (retry, retry_count < 3)
                  │
                  └──▶ FAILED          (retry_count ≥ 3)
```

---

## Resilience & Reliability

### Idempotency

| Layer | Mechanism |
|-------|-----------|
| HTTP | `Idempotency-Key` header read in handler |
| Use Case | `FindByIdempotencyKey` check before creation |
| Database | Partial unique index on `idempotency_key` |

Duplicate detection happens at the use case layer before any DB write. The DB constraint is the last line of defence against race conditions.

### Outbox Retry

The worker uses `MarkForRetry` — a single atomic SQL update:

```sql
UPDATE outbox
SET status      = CASE WHEN retry_count + 1 >= 3 THEN 'FAILED' ELSE 'PENDING' END,
    retry_count = retry_count + 1
WHERE id = $1
```

This means a failed publish attempt puts the event back to `PENDING` for the next poll cycle. After 3 failures the event is permanently marked `FAILED` for manual investigation.

### Concurrent Worker Safety

`FetchPending` uses `SELECT ... FOR UPDATE SKIP LOCKED` inside a transaction. Multiple worker replicas can run without processing the same event twice.

### HTTP Server Hardening

```go
server := &http.Server{
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}
// Graceful shutdown with 10 s deadline
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
```

### Connection Pool Tuning

```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
db.SetConnMaxIdleTime(1 * time.Minute)
```

---

## Observability

The full observability stack is included in `docker-compose.yml`:

| Tool | URL | Purpose |
|------|-----|---------|
| **Prometheus** | `localhost:9090` | Metrics scraping |
| **Grafana** | `localhost:3000` | Dashboards (admin/admin) |
| **Loki** | `localhost:3100` | Log aggregation |
| **Promtail** | — | Log shipping from Docker |

Metrics exposed by the service (via Prometheus middleware on every request):

- `http_requests_total` — counter by method, path, status
- `http_request_duration_seconds` — histogram of response latency
- Standard Go runtime metrics (goroutines, GC, memory)

---

## CI/CD Pipeline

Four GitHub Actions workflows run on every push to `main` and `staging`:

### Quality — `transaction-service-quality.yml`

| Job | What it does |
|-----|-------------|
| **Format** | Runs `gofmt` + `goimports`, auto-commits fixes back to the branch |
| **Vet & Build** | `go vet ./...` + full binary build |

### Lint — `transaction-service-lint.yml`

Runs `golangci-lint` with the project's `.golangci.yml` config.

Active linters: `errcheck`, `govet`, `staticcheck`, `gosimple`, `ineffassign`, `unused`, `bodyclose`, `misspell`, `revive`, `copyloopvar`.

### Tests — `transaction-service-tests.yml`

- Runs all tests with the **race detector** (`go test -race`)
- Enforces a **90% coverage gate** — the build fails below that threshold
- Uploads coverage report as a build artifact (7-day retention)

### Claude Code Review — `claude.yml`

On every pull request, Claude Code automatically:
- Reviews all changed files for correctness, Go idioms, architecture adherence, error handling, security, and performance
- Posts a PR review with inline comments on changed lines
- Prioritises the "Notes for @claude" section in the PR description

---

## Load Testing

k6 load tests are located in `transaction-service/tests/load/`. Three sequential scenarios enforce strict SLOs:

| Scenario | VUs | Duration | Goal |
|----------|-----|----------|------|
| **Smoke** | 1 | 30 s | Prove basic correctness before load |
| **Steady** | 0 → 50 → 0 | ~5 min | Validate p90 under normal traffic |
| **Spike** | 0 → 200 → 0 | ~50 s | Validate behaviour under burst traffic |

**SLO thresholds** (k6 fails the run if violated):

```
http_req_duration                p(90) < 30 ms
error_rate                       rate  < 1 %
duration_create_transaction      p(90) < 30 ms
duration_get_status              p(90) < 30 ms
duration_get_balance             p(90) < 30 ms
```

**Traffic mix** (realistic read-heavy distribution):
- 60% `GET /balance/{userId}`
- 30% `POST /transactions` (unique `Idempotency-Key` per request)
- 10% `GET /transactions/{id}`

### Run locally

```bash
# Install k6 — https://k6.io/docs/get-started/installation/
k6 run transaction-service/tests/load/load-test.js \
  -e BASE_URL=http://localhost:8080
```

### Run via Docker (no local k6 install needed)

```bash
# 1. Start the full stack
cd transaction-service
docker compose up -d

# 2. Run k6 inside the Docker network
docker compose -f docker-compose.yml -f docker-compose.load-test.yml run --rm k6
```

---

## Getting Started

### Prerequisites

- Docker + Docker Compose
- Go 1.22+ (for local development)

### Run everything

```bash
# Clone
git clone https://github.com/hebertzin/outbox-pattern.git
cd outbox-pattern/transaction-service

# Start all services (app, worker, postgres, rabbitmq, observability stack)
docker compose up --build -d

# Verify the API is up
curl http://localhost:8080/api/v1/balance/any-user-id
# → {"code":200,"message":"ok","data":{"user_id":"any-user-id","balance":0}}
```

### Run tests

```bash
cd transaction-service

# All tests
go test ./...

# With race detector + coverage
go test -race ./... -coverprofile=coverage.out
go tool cover -html=coverage.out   # open in browser

# Only use case layer
go test -v ./internal/core/usecase/...
```

### Run linter

```bash
cd transaction-service
golangci-lint run --config=.golangci.yml
```

### Apply database migrations

Migrations live in `transaction-service/migrations/`. Run them in order:

```bash
psql -h localhost -U postgres -d transaction_db \
  -f migrations/create_transaction_table.sql \
  -f migrations/create_outbox_table.sql \
  -f migrations/add_idempotency_key.sql \
  -f migrations/add_outbox_retry.sql
```

---

## Configuration

### Transaction Service

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | HTTP server port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASSWORD` | `postgres` | PostgreSQL password |
| `DB_NAME` | `transaction_db` | PostgreSQL database name |
| `RABBIT_URL` | `amqp://guest:guest@localhost:5672/` | RabbitMQ connection string |

### Users Service

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | HTTP server port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASSWORD` | `postgres` | PostgreSQL password |
| `DB_NAME` | `users_db` | PostgreSQL database name |
| `OUTBOX_WORKER_INTERVAL` | `5s` | How often the worker polls |

---

## Project Structure

```
outbox-pattern/
├── .github/
│   ├── pull_request_template.md   # PR template with "Notes for @claude" section
│   └── workflows/
│       ├── claude.yml                        # Automated PR review
│       ├── transaction-service-quality.yml   # Format + vet + build
│       ├── transaction-service-lint.yml      # golangci-lint
│       └── transaction-service-tests.yml     # Tests + 90% coverage gate
│
├── transaction-service/
│   ├── cmd/
│   │   ├── main.go                # HTTP server entrypoint
│   │   └── worker/main.go         # Outbox worker entrypoint
│   ├── config/config.go           # Env-based config loading
│   ├── docs/                      # Auto-generated Swagger docs
│   ├── infra/
│   │   ├── db/connection.go       # DB connection + pool config
│   │   └── repository/            # Postgres adapters
│   ├── internal/core/
│   │   ├── broker/                # RabbitMQ connection + publisher
│   │   ├── domain/
│   │   │   ├── entity/            # Transaction, Outbox — pure Go structs
│   │   │   └── ports/             # Repository + Publisher interfaces
│   │   ├── errors/                # *Exception typed error pattern
│   │   ├── handler/               # HTTP handlers + factory
│   │   └── usecase/               # Business logic + factory
│   ├── migrations/                # Ordered SQL migration files
│   ├── observability/             # Prometheus, Grafana, Loki configs
│   ├── tests/load/                # k6 load test scripts
│   ├── .golangci.yml              # Linter configuration
│   ├── docker-compose.yml         # Full stack (app + infra + observability)
│   └── docker-compose.load-test.yml  # k6 overlay for load testing
│
└── users-service/
    ├── main.go
    ├── config/
    ├── domain/
    ├── application/
    ├── infra/
    ├── presentation/
    └── migrations/
```

---

## Design Decisions

### Why two separate processes for HTTP and worker?

`cmd/main.go` serves HTTP; `cmd/worker/main.go` runs the outbox poller. This means the worker can be **scaled independently** from the API — you can run one HTTP pod and three worker pods if the outbox backlog grows.

### Why `FOR UPDATE SKIP LOCKED`?

Standard `SELECT ... WHERE status='PENDING'` with multiple workers causes them to race on the same rows. `SKIP LOCKED` lets each worker claim a non-overlapping set of rows in a single round-trip — no application-level locking, no duplicate processing.

### Why `event.ID` as RabbitMQ `MessageId` instead of a new UUID?

The outbox entry ID is stable across retries. Using it as the broker's `MessageId` lets consumers implement deduplication by tracking seen IDs — producing an **end-to-end idempotent pipeline**.

### Why partial unique index for `idempotency_key`?

```sql
CREATE UNIQUE INDEX uidx_transactions_idempotency_key
    ON transactions (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
```

Transactions without an idempotency key (the majority) are excluded from the index, keeping it small and fast. The index only enforces uniqueness where it matters.

### Why `usecase.Factory` and `handler.Factory`?

Without factories, `cmd/main.go` must know the concrete constructor signature of every use case. With factories:

```go
// Before — 4 lines, constructor signatures leaked into main
createUC  := usecase.NewCreateTransactionUseCase(txRepo)
statusUC  := usecase.NewGetTransactionStatusUseCase(txRepo)
balanceUC := usecase.NewGetBalanceUseCase(txRepo)
txHandler := handler.NewHandler(createUC, statusUC, balanceUC)

// After — 1 line, main only sees the factory
txHandler := handler.NewHandlerFactory(usecase.NewFactory(txRepo, logger))
```

Adding a new use case only requires changes inside `usecase/factory.go` — not in `main.go`.

### Why inject `*slog.Logger` into use cases instead of using a global?

Global loggers make tests noisy and make it impossible to configure per-request log levels. Injecting a `*slog.Logger` lets tests pass `io.Discard` and production code pass a structured JSON logger — same interface, different behaviour.

---

*Built to study distributed systems patterns used in production fintech systems.*

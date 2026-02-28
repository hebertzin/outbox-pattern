# Golang Outbox Pattern — Distributed Consistency Study

A production-grade implementation of the **Outbox Pattern** in Go, designed to guarantee **data consistency, reliability, and scalability** in microservices architectures.

This project demonstrates how modern distributed systems guarantee reliable event publishing without compromising database consistency.

---

## Overview

The system contains two independent services:

- **User Service** — manages user creation with outbox event emission
- **Transaction Service** — manages financial transactions with outbox event emission

Each service:
- Owns its own database
- Implements an `outbox` table
- Publishes events asynchronously via a background worker

The objective is to simulate a production-grade, event-driven architecture with strong consistency guarantees.

---

## The Dual-Write Problem

In distributed systems, this failure scenario is common:

```
Service writes to the database  ✓
Service tries to publish event  ✗  ← broker unavailable
System becomes inconsistent     ✗
```

The **Outbox Pattern** solves this.

---

## The Outbox Solution

Instead of publishing events directly to a broker:

1. Business operation is performed
2. Domain change **and** event record are stored in the **same database transaction**
3. A background worker polls the `outbox` table
4. Events are published to a message broker
5. After successful publishing, the event is marked as `PROCESSED`

```
BEGIN TRANSACTION
  INSERT INTO <domain_table> (...)
  INSERT INTO outbox (id, type, payload, status)
COMMIT

-- Background worker:
SELECT * FROM outbox WHERE status = 'PENDING'
Publish event to broker
UPDATE outbox SET status = 'PROCESSED'
```

**Guarantees:**
- Atomic persistence (no partial writes)
- No lost events
- Eventual consistency
- At-least-once delivery
- Retry capability

---

## Architecture

```
Client
  │
  ▼
HTTP API (Presentation Layer)
  │
  ▼
Use Case Layer (Application Layer)
  │
  ▼
Repository Layer (Infrastructure)
  │
  ├── domain table (users / transactions)
  └── outbox table
           │
           ▼
     Outbox Worker (background poller)
           │
           ▼
     Message Broker (Kafka / RabbitMQ / SQS)
```

Each service follows **Clean Architecture**:

```
domain/ports/         ← interfaces (dependency inversion)
application/usecase/  ← business logic (depends only on ports)
infra/                ← concrete implementations (DB, messaging, worker)
presentation/         ← HTTP handlers
```

---

## Tech Stack

- Go (Golang)
- PostgreSQL
- Clean Architecture
- Outbox Pattern
- Graceful Shutdown
- Kafka / RabbitMQ (pluggable via `EventPublisher` interface)

---

## User Service API

### Create User

```
POST /users
```

Request:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

Response:
```json
{
  "code": 201,
  "message": "user created successfully",
  "data": { "id": "uuid" }
}
```

---

## Transaction Service API

### Create Transaction

```
POST /api/v1/transactions
```

Request:
```json
{
  "amount": 100,
  "description": "transfer",
  "fromUserId": "user-1",
  "toUserId": "user-2"
}
```

Response:
```json
{
  "transactionId": "uuid",
  "status": "PENDING"
}
```

### Get Transaction Status

```
GET /api/v1/transactions/{id}
```

### Get Balance

```
GET /api/v1/balance/{userId}
```

---

## Environment Variables

### User Service
```
SERVER_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=users_db
OUTBOX_WORKER_INTERVAL=5s
```

### Transaction Service
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=transaction_db
```

---

## How to Run

```bash
docker-compose up --build
```

Then:
1. Create a user / transaction
2. Observe event written to outbox table
3. Watch worker publish event
4. Verify status transitions: `PENDING` → `PROCESSED`

---

## Consistency Guarantees

- Atomic write of domain data + outbox event
- Retry-safe event publishing
- At-least-once delivery
- Idempotency-ready design
- Failure recovery via polling worker

## Failure Scenarios Covered

- Broker unavailable
- Worker crash mid-processing
- Partial batch processing
- Database rollback simulation

---

## Core Concepts Practiced

- Outbox Pattern
- Event-Driven Architecture
- Eventual Consistency
- Distributed System Failure Handling
- Clean Architecture / Hexagonal Architecture
- Dependency Inversion Principle
- Worker-based asynchronous processing
- Transaction management in Go

---

## Future Improvements

- Kafka / RabbitMQ integration
- Idempotency keys
- Observability (metrics, tracing, structured logging)
- Retry and dead-letter queue support

---

## Purpose

Built as a distributed systems study project to deepen backend architecture knowledge and simulate production-grade reliability patterns used in fintech and high-scale systems.

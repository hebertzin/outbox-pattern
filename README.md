# Transaction Service with Outbox Pattern Implementation

A production-grade transaction service built in Go, designed to ensure **data consistency, reliability, and scalability** using the **Outbox Pattern**.

This project demonstrates how modern distributed systems guarantee reliable event publishing without compromising database consistency.

---

## Overview

This service is responsible for:

* Creating financial transactions
* Persisting transaction data atomically
* Writing integration events to an outbox table
* Ensuring reliable asynchronous event delivery via background workers

The implementation follows **clean architecture principles**, ensuring separation between domain, application, infrastructure, and presentation layers.

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
  ├── transactions table
  └── outbox table
           │
           ▼
     Outbox Worker
           │
           ▼
     Message Broker (Kafka / SQS / RabbitMQ)
```

---

## Outbox Pattern

The Outbox Pattern ensures **atomicity between database writes and event publishing**.

Instead of publishing events directly to a message broker, the service:

1. Saves the transaction in the database
2. Saves the event in the `outbox` table
3. Commits the transaction
4. A background worker publishes the event asynchronously

This guarantees that events are never lost, even in case of failures.

---

## Example Flow

```
BEGIN TRANSACTION

INSERT INTO transactions (...)

INSERT INTO outbox (
  id,
  type,
  payload,
  status
)

COMMIT
```

Worker:

```
SELECT * FROM outbox WHERE status = 'PENDING'

Publish event

UPDATE outbox SET status = 'PROCESSED'
```

---

## Tech Stack

* Go (Golang)
* PostgreSQL
* Gorilla Mux
* Swagger (OpenAPI)
* Clean Architecture
* Outbox Pattern
* Graceful Shutdown

---

## API

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

---

## Swagger Documentation

Available at:

```
http://localhost:8080/swagger/index.html
```

---

### transactions

Stores the transaction data.

### outbox

Stores integration events to be processed asynchronously.

---

## Why Outbox Pattern?

Without outbox:

```
Save in DB succeeds
Publish event fails
System becomes inconsistent
```

With outbox:

```
Save in DB succeeds
Save in outbox succeeds
Worker retries publishing until success
Consistency guaranteed
```

---

## Reliability Guarantees

* Atomic writes
* No event loss
* Retry support
* Eventually consistent architecture
* Failure-resilient event publishing

---

## Running the Service

### Environment variables

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=transaction_db
```

### Run

```
go run cmd/api/main.go
```

---

## Graceful Shutdown

The service supports graceful shutdown to ensure in-flight requests complete safely.

---

## Production-Ready Features

* Clean architecture
* Dependency injection
* Graceful shutdown
* OpenAPI documentation
* Outbox pattern implementation
* Transactional integrity

---

## Future Improvements

* Outbox worker implementation
* Kafka / RabbitMQ integration
* Idempotency support
* Observability (metrics, tracing)
* Retry and dead-letter queue support

---

## Purpose

This project was built to demonstrate production-grade patterns used in modern distributed systems and fintech architectures.

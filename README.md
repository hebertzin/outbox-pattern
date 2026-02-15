# Golang Outbox Pattern - Distributed Consistency Study

> Production-inspired implementation of the **Outbox Pattern** in Golang
> to guarantee reliable event publishing in a microservices
> architecture.

------------------------------------------------------------------------

## Overview

This project demonstrates a practical implementation of the **Outbox
Pattern** to solve a classic distributed systems problem:

> How do we guarantee that database state changes and event publishing
> happen reliably and consistently?

The system currently contains two independent services:

-   **User Service**
-   **Transaction Service**

Each service:

-   Owns its database
-   Implements an `outbox` table
-   Publishes events asynchronously via a worker

The objective is to simulate a production-grade, event-driven
architecture with strong consistency guarantees.

------------------------------------------------------------------------

##  Problem Being Solved

In distributed systems, this failure scenario is common:

1.  Service writes to the database 
2.  Service tries to publish an event 
3.  System becomes inconsistent

This is known as the **dual-write problem**.

------------------------------------------------------------------------

## The Outbox Solution

Instead of publishing events directly:

1.  The service performs a business operation.
2.  The domain change and event record are stored **in the same database
    transaction**.
3.  A background worker polls the `outbox` table.
4.  Events are published to a message broker.
5.  After successful publishing, the event is marked as `PROCESSED`.

This guarantees:

-   Atomic persistence
-   No lost events
-   Eventual consistency
-   Retry capability

------------------------------------------------------------------------

##  Tech Stack

-   Golang
-   PostgreSQL
-   Worker Pattern
-   Kafka / RabbitMQ (planned)
-   Clean Architecture principles

------------------------------------------------------------------------

## Consistency Guarantees

This implementation provides:

-   Atomic write of domain + event
-   Retry-safe event publishing
-   At-least-once delivery
-   Idempotency-ready design
-   Failure recovery via polling worker

------------------------------------------------------------------------

## Core Concepts Practiced

-   Outbox Pattern
-   Event-Driven Architecture
-   Eventual Consistency
-   Distributed System Failure Handling
-   Idempotency
-   Worker-based asynchronous processing
-   Transaction management in Go

------------------------------------------------------------------------

## Failure Scenarios Covered

-   Broker unavailable
-   Worker crash
-   Partial processing
-   Duplicate event handling
-   Transaction rollback simulation

------------------------------------------------------------------------

##  Why This Project Matters

This pattern is widely used in high-scale systems to guarantee
consistency between state changes and asynchronous messaging.

Understanding and implementing this correctly is critical for backend
engineers working in distributed systems.

------------------------------------------------------------------------

##  How to Run

``` bash
docker-compose up --build
```

Then:

1.  Create a user
2.  Observe event being written to outbox
3.  Watch worker publish event
4.  Verify status transition

------------------------------------------------------------------------

## Author

Built as a distributed systems study project to deepen backend
architecture knowledge and simulate production-grade reliability
patterns.

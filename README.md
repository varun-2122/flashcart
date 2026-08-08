# 🚀 FlashCart - Production-Grade Backend Engine

> **High-Performance, Production-Grade E-Commerce Backend written in Go, implementing Clean Architecture, Domain-Driven Design (DDD), and Event-Driven Asynchronous Processing.**

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Architecture](https://img.shields.io/badge/Architecture-Clean%20%2F%20DDD-orange)](#-1-system-architecture)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## 💡 Overview & Engineering Objectives

FlashCart is a **production-grade e-commerce backend system** written in **Go**. Built for high throughput, strict request execution bounds, resilient database pooling, and observable event processing, FlashCart simulates how companies like **Google, Uber, Amazon, and Rubrik** design distributed backend services.

---

## 📐 1. System Architecture Block Diagram

Our backend architecture separates concerns across isolated architectural layers: Client → NGINX Load Balancer / API Gateway → HTTP Middleware Chain → Service Layer → Data Repositories → PostgreSQL & Redis Cache.

![FlashCart System Block Diagram](docs/architecture/system_block_diagram.png)

---

## 🎯 2. Use Case Diagram

Detailed functional decomposition of FlashCart actors and domain capabilities across Auth, Catalog, Cart, Orders, and Stock Reservations.

![FlashCart Use Case Diagram](docs/architecture/use_case_diagram.png)

---

## 🔄 3. Chronological Request Operations Trace (Sequence)

How incoming requests execute through FlashCart's zero-panic middleware pipeline, transactional Unit of Work database locks, and background async worker pools.

![FlashCart Request Lifecycle Diagram](docs/architecture/request_lifecycle_diagram.png)

```text
Client (HTTP POST /orders)
  │
  ├──► [ Middleware Layer ]
  │      ├── Request ID Generator (X-Request-ID)
  │      ├── Recovery Handler (Panic Safety)
  │      ├── Timeout Context (5s Deadline)
  │      └── Structured Logger (slog JSON)
  │
  ├──► [ HTTP Handler & Service Layer ]
  │      ├── Input Validation & DTO Decoding
  │      └── Business Domain Rules Execution
  │
  ├──► [ Data Layer - PostgreSQL DB Transaction ]
  │      ├── BEGIN TX (pgxpool)
  │      ├── Reserve Inventory (Optimistic Locking: WHERE version = x)
  │      ├── Create Order Entity
  │      └── COMMIT TX
  │
  ├──► [ Async Worker Pool ]
  │      ├── Dispatch Notification / Email
  │      └── Trigger Analytics & Invoice Jobs
  │
  └──► HTTP 201 Created Response
```

---

## 🛠 Tech Stack

- **Backend**: Go 1.26+ (Standard Library, `net/http`)
- **Database**: PostgreSQL 16 (`jackc/pgx/v5` with connection pooling via `pgxpool`)
- **Cache**: Redis 7 (`redis/go-redis/v9`)
- **Containerization**: Docker, Docker Compose
- **DevOps**: GitHub Actions CI/CD Pipeline
- **Observability**: Prometheus (`/metrics`), Grafana, OpenTelemetry, `slog` JSON Logging

---

## 📂 Clean Architecture Directory Structure

```text
flashcart/
├── cmd/
│   └── server/
│       └── main.go                 # App bootstrap & Graceful Shutdown listener
├── internal/
│   ├── config/config.go            # Environment configuration & timeout controls
│   ├── logger/logger.go            # Structured JSON logger with Context Trace IDs
│   ├── database/postgres.go        # PostgreSQL connection pool (pgxpool)
│   ├── cache/redis.go              # Redis cache connection manager
│   ├── response/response.go        # Standardized HTTP JSON API responses
│   ├── middleware/middleware.go    # Request ID, Recovery, Logger, Timeout, CORS
│   ├── server/
│   │   ├── server.go               # HTTP server setup & lifecycle
│   │   └── health.go               # Health, Liveness, and Readiness probes
│   ├── domain/                     # Core enterprise domain entities & interfaces
│   ├── auth/                       # Authentication & JWT RBAC module skeleton
│   ├── user/                       # User management skeleton
│   ├── product/                    # Product catalog skeleton
│   ├── inventory/                  # Inventory reservation & optimistic locking skeleton
│   ├── cart/                       # Shopping cart skeleton
│   ├── order/                      # Order state machine & DB transactions skeleton
│   ├── payment/                    # Payment gateway integration skeleton
│   ├── shipping/                   # Logistics & shipment tracking skeleton
│   └── notification/               # Asynchronous task notification skeleton
├── docs/
│   └── architecture/               # Excalidraw architecture & sequence diagrams
├── deploy/                         # Production Kubernetes & Docker assets
├── docker-compose.yml              # Local container environment (PostgreSQL + Redis + App)
├── Dockerfile                      # Multi-stage production container build
├── Makefile                        # Automation commands
├── .env.example                    # Environment variable template
└── go.mod
```

---

## 🚦 Engineering Phase Roadmap

| Phase | Description | Status |
|---|---|---|
| **Phase 1** | Core Infrastructure Foundation (Clean Architecture, Config, Structured `slog` Logger, Postgres `pgxpool`, Redis Client, Health Probes, Graceful Teardown, Docker Compose) | ✅ **Completed** |
| **Phase 2** | Domain Engineering & Core Modules (JWT Auth, Refresh Tokens, Product Catalog, Inventory Optimistic Locking, Order DB Transactions) | ⏳ Next |
| **Phase 3** | Concurrency Engine & Business Observability (Go Worker Pools, Prometheus `/metrics`, Grafana Business Dashboard) | 🔜 Planned |
| **Phase 4** | Production Observability & Deployment (OpenTelemetry Tracing, GitHub Actions CI/CD Pipeline, Kubernetes & Helm Deployment) | 🔜 Planned |

---

## 🚀 Deployment Guide

### Option 1: Local Docker Compose Deployment (Recommended)

Spins up PostgreSQL 16, Redis 7, and FlashCart API container with health checks and volume persistence:

```bash
# 1. Clone repository
git clone https://github.com/varun-2122/flashcart.git
cd flashcart

# 2. Copy environment configuration
cp .env.example .env

# 3. Launch container stack
make docker-up
```

Verify services:
- API Server: `http://localhost:8080`
- Liveness Probe: `http://localhost:8080/healthz`
- Deep Readiness Probe: `http://localhost:8080/readyz`

```bash
# Stop container stack
make docker-down
```

---

### Option 2: Production Multi-Stage Docker Deployment

Build a minimal, secure, non-root Alpine container:

```bash
docker build -t flashcart:v1.0.0 .
docker run -p 8080:8080 --env-file .env flashcart:v1.0.0
```

---

### Option 3: Direct Host Binary Execution

```bash
make build
./bin/server
```

---

## 📊 Health Probes API Reference

| Endpoint | Method | Description | Success Response | Status |
|---|---|---|---|---|
| `/` | `GET` | API Engine Information | `{"name":"FlashCart API Engine","version":"v1.0.0","status":"operational"}` | `200 OK` |
| `/healthz` | `GET` | System Liveness Probe | `{"success":true,"data":{"status":"UP"}}` | `200 OK` |
| `/livez` | `GET` | Kubernetes Liveness Probe | `{"success":true,"data":{"status":"ALIVE"}}` | `200 OK` |
| `/readyz` | `GET` | Deep Readiness Probe (DB + Redis) | `{"status":"READY","database":"UP","cache":"UP"}` | `200 OK` / `503 Service Unavailable` |

---

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

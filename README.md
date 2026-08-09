# 🚀 FlashCart - Production-Grade Backend Engine

> **High-Performance, Production-Grade E-Commerce Backend written in Go, implementing Clean Architecture, Domain-Driven Design (DDD), and Event-Driven Asynchronous Processing.**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Architecture](https://img.shields.io/badge/Architecture-Clean%20%2F%20DDD-orange)](#-1-system-architecture-block-diagram)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## 💡 Overview & Engineering Objectives

FlashCart is a **production-grade e-commerce backend system** written in **Go**. Built for high throughput, strict request execution bounds, resilient database pooling, and observable event processing, FlashCart simulates how modern engineering teams at companies like **Google, Uber, Amazon, and Rubrik** design distributed backend services.

### Core Engineering Features
- **Clean Architecture & DDD**: Strict layer isolation (Client → Middleware → Handler → Service → Repository → Database/Cache).
- **High Concurrency & Low Latency**: Native Go `net/http` multiplexing with zero-allocation `log/slog` JSON logging.
- **Resilient Connection Pooling**: PostgreSQL connection management via `pgx/v5` (`pgxpool`) and Redis caching via `go-redis/v9`.
- **Fault-Tolerant Middleware Pipeline**:
  - `RequestID`: Trace propagation with `X-Request-ID`.
  - `Recovery`: Panic safety recovering from runtime panics without dropping server execution.
  - `Timeout`: Strict request execution boundary via `context.WithTimeout`.
  - `Logger`: Structured JSON metrics (latency, HTTP status, client IP, payload bytes).
- **Optimistic Concurrency Control**: Stock reservation with versioned database locks preventing race conditions during flash sales.
- **Unit of Work Transactions**: Multi-table order processing inside atomic PostgreSQL transactions (`pgx.Tx`).
- **Observability Probes**: System liveness (`/livez`), health (`/healthz`), and deep readiness (`/readyz`) endpoints.

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

## 🛠 4. Tech Stack

- **Language & Runtime**: Go 1.24+ (Standard Library, `net/http`)
- **Database**: PostgreSQL 16 (`jackc/pgx/v5` connection pooling via `pgxpool`)
- **Cache**: Redis 7 (`redis/go-redis/v9`)
- **Security**: JWT (`golang-jwt/jwt/v5`), Bcrypt password hashing (`golang.org/x/crypto`)
- **Containerization & Dev Tooling**: Docker, Docker Compose, Makefile
- **DevOps**: GitHub Actions CI/CD Pipeline
- **Observability**: Prometheus (`/metrics`), Grafana, OpenTelemetry, `slog` JSON Logging

---

## 📂 5. Clean Architecture Directory Structure

```text
flashcart/
├── cmd/
│   └── server/
│       └── main.go                 # App bootstrap & Graceful Shutdown listener
├── internal/
│   ├── config/config.go            # Environment configuration & timeout controls
│   ├── logger/logger.go            # Structured JSON logger with Context Trace IDs
│   ├── database/                   # PostgreSQL connection pool & schema migrations
│   │   ├── postgres.go
│   │   └── migrations.go
│   ├── cache/redis.go              # Redis cache connection manager
│   ├── response/response.go        # Standardized HTTP JSON API responses
│   ├── middleware/middleware.go    # Request ID, Recovery, Logger, Timeout, CORS
│   ├── server/
│   │   ├── server.go               # HTTP server setup & domain route wiring
│   │   └── health.go               # Health, Liveness, and Readiness probes
│   ├── domain/                     # Core enterprise domain entities & interfaces
│   │   ├── user.go
│   │   ├── product.go
│   │   ├── inventory.go
│   │   ├── cart.go
│   │   └── order.go
│   ├── auth/                       # JWT Authentication, Bcrypt, & RBAC Middleware
│   ├── user/                       # User profile persistence
│   ├── product/                    # Product catalog & Redis cache decorator
│   ├── inventory/                  # Inventory optimistic locking stock manager
│   ├── cart/                       # Redis shopping cart persistence
│   ├── order/                      # Order checkout engine & Unit of Work transactions
│   ├── payment/                    # Payment gateway integration skeleton
│   ├── shipping/                   # Logistics & shipment tracking skeleton
│   └── notification/               # Asynchronous task notification skeleton
├── scripts/
│   └── migrations/                 # DDL SQL Schema Migration Scripts
├── docs/
│   └── architecture/               # Excalidraw architecture & sequence diagrams
├── docker-compose.yml              # Local container environment (PostgreSQL + Redis + App)
├── Dockerfile                      # Multi-stage production container build
├── Makefile                        # Automation commands
├── .env.example                    # Environment variable template
└── go.mod
```

---

## ⚡ 6. Local Development & Setup

### Prerequisites
- Go 1.24+ installed
- Docker & Docker Compose (Optional for local PostgreSQL and Redis containers)

### Quick Start

1. **Clone the Repository**
   ```bash
   git clone https://github.com/varun-2122/flashcart.git
   cd flashcart
   ```

2. **Configure Environment Variables**
   ```bash
   cp .env.example .env
   ```

3. **Start Storage Services (PostgreSQL & Redis)**
   ```bash
   # Using Docker Compose
   make docker-up
   ```

4. **Run the Backend Engine**
   ```bash
   make run
   ```

5. **Execute Unit Test Suite**
   ```bash
   make test
   ```

---

## 🚦 7. Engineering Phase Roadmap

| Phase | Description | Status |
|---|---|---|
| **Phase 1** | Core Infrastructure Foundation (Clean Architecture, Config, Structured `slog` Logger, Postgres `pgxpool`, Redis Client, Health Probes, Graceful Teardown, Docker Compose) | ✅ **Completed** |
| **Phase 2** | Domain Engineering & Core Modules (JWT Auth, Refresh Tokens, Product Catalog, Inventory Optimistic Locking, Cart, Order DB Transactions) | ✅ **Completed** |
| **Phase 3** | Concurrency Engine & Business Observability (Go Worker Pools, Prometheus `/metrics`, Grafana Business Dashboard) | ⏳ **Next** |
| **Phase 4** | Production Observability & Cloud Native (OpenTelemetry Tracing, GitHub Actions CI/CD Pipeline, Kubernetes & Helm Deployment) | 🔜 Planned |

---

## 📊 8. Endpoints & API Reference

### System & Health Probes

| Endpoint | Method | Description | Success Response | Status |
|---|---|---|---|---|
| `/` | `GET` | API Engine Information | `{"name":"FlashCart API Engine","version":"v2.0.0","status":"operational"}` | `200 OK` |
| `/healthz` | `GET` | System Liveness Probe | `{"success":true,"data":{"status":"UP"}}` | `200 OK` |
| `/livez` | `GET` | Kubernetes Liveness Probe | `{"success":true,"data":{"status":"ALIVE"}}` | `200 OK` |
| `/readyz` | `GET` | Deep Readiness Probe (DB + Redis) | `{"status":"READY","database":"UP","cache":"UP"}` | `200 OK` / `503 Service Unavailable` |

### Auth API

| Endpoint | Method | Auth | Description | Payload / Query |
|---|---|---|---|---|
| `/api/v1/auth/register` | `POST` | None | Register new user account | `{"email", "password", "first_name", "last_name", "role"}` |
| `/api/v1/auth/login` | `POST` | None | Authenticate and obtain JWT | `{"email", "password"}` |

### Product Catalog API

| Endpoint | Method | Auth | Description | Payload / Query |
|---|---|---|---|---|
| `/api/v1/products` | `GET` | None | List catalog with pagination & search | `?search=...&category_id=...&min_price=...&limit=20&offset=0` |
| `/api/v1/products/{id}` | `GET` | None | Get product by UUID (Redis cached) | Path Param `{id}` |
| `/api/v1/products` | `POST` | Admin | Create product & initial stock | `{"sku", "name", "price", "brand", "quantity"}` |

### Shopping Cart & Order API

| Endpoint | Method | Auth | Description | Payload / Query |
|---|---|---|---|---|
| `/api/v1/cart` | `GET` | Bearer JWT | View user shopping cart | Header `Authorization: Bearer <token>` |
| `/api/v1/cart/items` | `POST` | Bearer JWT | Add item to cart | `{"product_id", "quantity"}` |
| `/api/v1/cart/items/{product_id}` | `DELETE` | Bearer JWT | Remove item from cart | Path Param `{product_id}` |
| `/api/v1/orders` | `POST` | Bearer JWT | Process checkout (Unit of Work DB Tx) | Header `Authorization: Bearer <token>` |
| `/api/v1/orders` | `GET` | Bearer JWT | List authenticated user orders | Header `Authorization: Bearer <token>` |

---

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

# 🚀 FlashCart - Production-Grade Backend Engine

> **High-Performance, Production-Grade E-Commerce Backend written in Go, implementing Clean Architecture, Domain-Driven Design (DDD), and Event-Driven Asynchronous Processing.**

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Architecture](https://img.shields.io/badge/Architecture-Clean%20%2F%20DDD-orange)](#-architecture)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## 💡 Overview

FlashCart is a **production-grade backend engine** engineered to demonstrate top-tier software engineering standards (Google / Uber / Atlassian / Rubrik level). Unlike generic CRUD applications, FlashCart is designed to handle high concurrency, strict request bounds, observable telemetry, resilient caching, connection pooling, and fault-tolerant background execution.

### Key Highlights
- **Clean Architecture & Domain-Driven Design (DDD)**: Strict separation of concerns (Client → API Gateway → Middleware → Handler → Service → Repository → Database/Cache).
- **High Concurrency & Low Latency**: Native Go `net/http` multiplexing with zero-allocation `log/slog` JSON logging.
- **Resilient Connection Pooling**: PostgreSQL connection management using `pgx/v5` (`pgxpool`) and Redis cache management using `go-redis/v9`.
- **Fault-Tolerant HTTP Middleware**:
  - `RequestID`: Trace propagation with `X-Request-ID`.
  - `Recovery`: Panic safety recovering from runtime panics without dropping server execution.
  - `Timeout`: Strict 5-second request execution boundary via `context.WithTimeout`.
  - `Logger`: Structured JSON metrics (latency, status, client IP, payload bytes).
- **Observability Probes**: Production liveness (`/livez`), health (`/healthz`), and deep readiness (`/readyz`) endpoints.
- **Graceful Lifecycle Management**: Clean OS signal trapping (`SIGINT`, `SIGTERM`) with orderly teardown of database pools and HTTP server listeners.

---

## 🛠 Tech Stack

- **Language**: Go 1.26+ (Standard Library)
- **Database**: PostgreSQL 16 (`pgxpool`)
- **Cache**: Redis 7 (`go-redis/v9`)
- **Containerization**: Docker, Docker Compose
- **Configuration**: `godotenv` with Environment Variable Overrides

---

## 📂 Project Structure

```text
flashcart/
├── cmd/
│   └── server/
│       └── main.go                 # Application bootstrap & Graceful Shutdown listener
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
│   ├── auth/                       # Authentication & JWT RBAC skeleton
│   ├── user/                       # User management skeleton
│   ├── product/                    # Product catalog skeleton
│   ├── inventory/                  # Inventory reservation & optimistic locking skeleton
│   ├── cart/                       # Shopping cart skeleton
│   ├── order/                      # Order state machine & DB transactions skeleton
│   ├── payment/                    # Payment gateway integration skeleton
│   ├── shipping/                   # Logistics & shipment tracking skeleton
│   └── notification/               # Asynchronous task notification skeleton
├── deploy/                         # Deployment manifests
├── docker-compose.yml              # Local container environment (PostgreSQL + Redis + App)
├── Dockerfile                      # Multi-stage production container build
├── Makefile                        # Automation commands
├── .env.example                    # Environment variable template
└── go.mod
```

---

## 🚦 Phase Roadmap

| Phase | Description | Status |
|---|---|---|
| **Phase 1** | Core Foundation & Infrastructure (Clean Arch, Config, Logger, Postgres, Redis, Health Probes, Graceful Shutdown, Docker) | ✅ **Completed** |
| **Phase 2** | Domain Engineering & Clean Arch Core (Auth, User, Product, Inventory Optimistic Locking, Order DB Tx) | ⏳ Next |
| **Phase 3** | Concurrency Engine & Business Observability (Go Worker Pools, Prometheus `/metrics`, Grafana Dashboards) | 🔜 Planned |
| **Phase 4** | Enterprise Observability & Cloud Native (OpenTelemetry Distributed Tracing, CI/CD, K8s, Helm) | 🔜 Planned |

---

## ⚡ Quick Start

### Prerequisites
- Go 1.24+ installed
- Docker & Docker Compose installed

### Local Development

1. **Clone the Repository**
   ```bash
   git clone https://github.com/varun-2122/flashcart.git
   cd flashcart
   ```

2. **Set Up Environment Variables**
   ```bash
   cp .env.example .env
   ```

3. **Start Local Dependencies with Docker**
   ```bash
   make docker-up
   ```

4. **Run the Application**
   ```bash
   make run
   ```

5. **Run Unit Tests**
   ```bash
   make test
   ```

---

## 📊 Health Probes

| Route | Method | Description | Success Response |
|---|---|---|---|
| `/` | `GET` | API Engine Meta Info | `{"name":"FlashCart API Engine","version":"v1.0.0","status":"operational"}` |
| `/healthz` | `GET` | System Liveness Probe | `{"success":true,"data":{"status":"UP"}}` |
| `/livez` | `GET` | K8s Liveness Check | `{"success":true,"data":{"status":"ALIVE"}}` |
| `/readyz` | `GET` | Deep Readiness Probe (Postgres + Redis) | `{"status":"READY","database":"UP","cache":"UP"}` |

---

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

# 🚀 FlashCart - Production-Grade E-Commerce Backend Engine

> **High-Performance, Production-Grade E-Commerce Backend written in Go, implementing Clean Architecture, Domain-Driven Design (DDD), and Event-Driven Asynchronous Processing.**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Deployment Status](https://img.shields.io/badge/Deployment-Live%20on%20Render-brightgreen?style=flat&logo=render)](https://flashcart-api-bob1.onrender.com)
[![Architecture](https://img.shields.io/badge/Architecture-Clean%20%2F%20DDD-orange)](#-1-system-architecture)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## 🌐 Live Production API

- 🔗 **API Base Engine**: [https://flashcart-api-bob1.onrender.com](https://flashcart-api-bob1.onrender.com)
- 💚 **Liveness Probe**: [https://flashcart-api-bob1.onrender.com/healthz](https://flashcart-api-bob1.onrender.com/healthz)
- ⚡ **Readiness Probe**: [https://flashcart-api-bob1.onrender.com/readyz](https://flashcart-api-bob1.onrender.com/readyz)

---

## 💡 Engineering Highlights

FlashCart is built to demonstrate production backend engineering standards (Google / Uber / Atlassian / Rubrik level):

- **Clean Architecture & DDD**: Strict layer isolation (Client → Middleware → Handler → Service → Repository → Database/Cache).
- **JWT & RBAC Security**: HMAC-SHA256 tokens, Bcrypt password hashing, and Role-Based Access Control (`customer`, `admin`).
- **High-Concurrency Optimistic Locking**: Stock reservation using versioned database updates (`WHERE product_id = $1 AND version = $2`).
- **Unit of Work Transactions**: Atomic multi-table order checkout executed inside PostgreSQL transactions (`pgx.Tx`).
- **Redis Catalog Caching**: Transparent cache decorator for hot product queries (`product:{id}`) with automatic invalidation.
- **Zero-Panic Middleware**: `X-Request-ID` tracing, `slog` JSON logging, panic recovery, and strict 5s request context deadlines.

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
  ├──► [ Middleware Layer ] (Request ID, 5s Context Timeout, slog Logger, Panic Recovery)
  ├──► [ HTTP Handler & Service Layer ] (Validation, Auth Claims, DTO Parsing)
  ├──► [ Data Layer - PostgreSQL DB Transaction ]
  │      ├── BEGIN TX (pgxpool)
  │      ├── Reserve Inventory (Optimistic Locking: WHERE version = x)
  │      ├── Create Order & OrderItems Entity
  │      └── COMMIT TX
  ├──► [ Async Operations ] (Clear Cart, Dispatch Notifications)
  └──► HTTP 201 Created Response
```

---

## 🛠 Tech Stack

- **Backend**: Go 1.24+ (`net/http`, `log/slog`)
- **Database**: PostgreSQL 16 (`jackc/pgx/v5` with connection pooling via `pgxpool`)
- **Cache**: Redis 7 (`redis/go-redis/v9`)
- **Authentication**: JWT (`golang-jwt/jwt/v5`), Bcrypt (`golang.org/x/crypto`)
- **DevOps**: Docker, Docker Compose, GitHub Actions CI/CD Pipeline

---

## 📂 Clean Architecture Directory Structure

```text
flashcart/
├── cmd/
│   └── server/
│       └── main.go                 # App bootstrap & Graceful Shutdown listener
├── internal/
│   ├── auth/                       # JWT generation, Bcrypt, AuthService, RBAC Middleware
│   ├── cart/                       # Redis Shopping Cart repository, service, & handlers
│   ├── inventory/                  # Optimistic Locking stock reservation repository
│   ├── order/                      # Unit of Work DB Transaction order checkout
│   ├── product/                    # Product catalog repository, Redis cache decorator, & service
│   ├── user/                       # User profile repository & domain logic
│   ├── config/config.go            # Environment configuration & timeout controls
│   ├── logger/logger.go            # Structured slog JSON logger with Trace IDs
│   ├── database/                   # PostgreSQL pgxpool manager & auto-migrations
│   ├── cache/redis.go              # Redis cache connection manager
│   ├── response/response.go        # Standardized HTTP JSON API response helper
│   ├── middleware/middleware.go    # Request ID, Recovery, Logger, Timeout, CORS
│   ├── server/                     # HTTP server setup & Health/Readiness probes
│   └── domain/                     # Core enterprise domain entities & interfaces
├── scripts/
│   └── migrations/                 # DDL SQL Schema Migration Scripts
├── docs/
│   └── architecture/               # Excalidraw architecture & sequence diagrams
├── render.yaml                     # Infrastructure as Code (1-Click Cloud Blueprint)
├── docker-compose.yml              # Local container environment (PostgreSQL + Redis + App)
├── Dockerfile                      # Multi-stage production container build
├── Makefile                        # Automation commands
└── go.mod
```

---

## 🔌 API Endpoints Reference

### 🔐 Authentication (`/api/v1/auth`)
| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `POST` | `/api/v1/auth/register` | Register new user | No |
| `POST` | `/api/v1/auth/login` | Authenticate user & receive JWT token | No |

### 🛍 Products (`/api/v1/products`)
| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `GET` | `/api/v1/products` | List, filter, search & paginate products | No |
| `GET` | `/api/v1/products/{id}` | Get product details by UUID | No |
| `POST` | `/api/v1/products` | Create product with stock | **Admin Only** |

### 🛒 Shopping Cart (`/api/v1/cart`)
| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `GET` | `/api/v1/cart` | Get current user's shopping cart | **Yes** |
| `POST` | `/api/v1/cart/items` | Add or update item quantity in cart | **Yes** |
| `DELETE` | `/api/v1/cart/items/{product_id}` | Remove item from cart | **Yes** |

### 📦 Orders (`/api/v1/orders`)
| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `POST` | `/api/v1/orders` | Checkout cart with **Optimistic Stock Locking** | **Yes** |
| `GET` | `/api/v1/orders` | List current user's order history | **Yes** |
| `GET` | `/api/v1/orders/{id}` | Get order details by UUID | **Yes** |

### 📊 Health & Telemetry Probes
| Method | Endpoint | Description | Response |
|---|---|---|---|
| `GET` | `/healthz` | System Liveness Probe | `{"success":true,"data":{"status":"UP"}}` |
| `GET` | `/readyz` | Deep Readiness Probe (Postgres + Redis) | `{"status":"READY","database":"UP","cache":"UP"}` |

---

## ⚡ Quick Start

```bash
# 1. Clone repository
git clone https://github.com/varun-2122/flashcart.git
cd flashcart

# 2. Build and run server binary
make build
./bin/server
```

---

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

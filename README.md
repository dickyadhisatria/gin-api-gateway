# Gin API Gateway

A robust, secure, and high-performance API Gateway built from scratch using **Go** and the **Gin Gonic** framework. This project is engineered with architectural precision using **Clean Architecture** principles, enforcing strict separation of concerns, the dependency rule, and dependency inversion.

It serves as the single entry point for microservices architectures, managing traffic routing, security, and system resilience.

## 🚀 Features

- **Clean Architecture Implementation**: Strict division into Domain, Use Cases, Adapters, and Infrastructure layers. Completely decoupled from specific frameworks for testability.
- **Reverse Proxy Routing**: Dynamically routes incoming requests to downstream microservices.
- **Tuned Connection Pooling**: Configured `http.Transport` settings (`MaxIdleConnsPerHost: 100`) to prevent TIME_WAIT socket exhaustion and connection establishment latency.
- **Distributed Rate Limiting**: Supports both in-memory and **Redis** backends using `ulule/limiter` to scale horizontally. Falls back to in-memory if no Redis is configured.
- **Security & Authentication**: Implements JWT-based authentication and authorization for secure access to microservices.
- **Fault Tolerance (Circuit Breaker)**: Prevents cascading failures using `sony/gobreaker` dynamically mapped and cached per destination host.
- **Downstream Timeout Guard**: Enforces a 5-second context timeout on downstream microservices to avoid goroutine leaks.
- **Configuration Management**: Zero hardcoded values, completely driven by environment variables (`.env`).
- **Logging & Monitoring**: Structured JSON request logging (`slog`) with cryptographically secure Request IDs and customized JSON gateway errors.

## 📁 Project Structure

```text
gin-api-gateway/
├── cmd/
│   └── api/
│       └── main.go                 # Composition Root (Dependency Injection)
├── internal/
│   ├── domain/
│   │   ├── model/                  # Domain entities (e.g. user claims)
│   │   └── service/                # Domain interface definitions
│   ├── usecase/                    # Business use cases (e.g. ProxyUseCase)
│   ├── adapter/
│   │   ├── auth/                   # JWT verification adapter
│   │   ├── ratelimit/              # Memory & Redis rate limiter adapters
│   │   ├── proxy/                  # httputil + gobreaker reverse proxy adapter
│   │   └── delivery/
│   │       └── http/
│   │           ├── handler/        # HTTP controllers/handlers
│   │           └── middleware/     # HTTP logging, rate limiting, and auth middlewares
│   └── infrastructure/
│       └── config/                 # Environment loader
├── .env                            # Local secret configuration (Git ignored)
├── .gitignore                      # Tells Git which files to ignore
├── go.mod                          # Go module configuration
├── go.sum                          # Go module checksums
├── k6.js                           # K6 load testing script
└── README.md                       # Documentation
```

## 🛠️ Getting Started

### 1. Prerequisites

Make sure you have **Go 1.26+** installed on your machine.

### 2. Installation

Clone the repository and download the required dependencies:

```bash
git clone https://github.com/dickyadhisatria/gin-api-gateway
cd gin-api-gateway
go mod tidy
```

### 3. Environment Setup

Create a `.env` file in the root directory:

```text
PORT=:3000
JWT_SECRET=your-secure-jwt-secret
AUTH_SERVICE_URL=http://localhost:3001
USER_SERVICE_URL=http://localhost:3002
PRODUCT_SERVICE_URL=http://localhost:3003
ORDER_SERVICE_URL=http://localhost:3004

# Redis Configuration (Optional, falls back to memory if empty)
REDIS_URL=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

### 4. Running the Application

Start the API Gateway:

```bash
go run cmd/api/main.go
```

The gateway will start listening on port `:3000`.

### 5. Running Tests

Execute unit tests:

```bash
go test -v ./...
```

## 🧪 Testing the Gateway

Test the JWT authentication using `cURL`:

```bash
# This will return 401 Unauthorized
curl -i http://localhost:3000/api/v1/user/profile

# This will successfully proxy to your user service
curl -i -H "Authorization: Bearer <your-jwt-token>" http://localhost:3000/api/v1/user/profile
```

# Irrigation Analytics API

A production-ready REST API for managing irrigation analytics in agricultural platforms, built with Go using Gin, GORM, and comprehensive observability tooling.

## Quick Start

### Prerequisites
- Go 1.22+
- Docker & Docker Compose

### Setup & Run

```bash
# 1. Start infrastructure services
docker-compose up -d

# 2. Install dependencies
go mod download

# 3. Run the API server
go run main.go
```

The API will start on `http://localhost:8080`

**Test the health endpoint:**
```bash
curl http://localhost:8080/health
```

### Access Services

| Service | URL | Purpose |
|---------|-----|---------|
| **API** | http://localhost:8080 | REST API |
| **Jaeger UI** | http://localhost:16686 | Distributed tracing visualization |
| **Grafana** | http://localhost:3000 | Log/metric dashboards (admin/admin) |
| **Loki** | http://localhost:3100 | Log aggregation |
| **PostgreSQL** | localhost:5432 | Database |

## Tech Stack

- **Framework:** [Gin](https://gin-gonic.com/) — Lightweight HTTP framework
- **ORM:** [GORM](https://gorm.io/) — Object-relational mapping with PostgreSQL
- **Logging:** [Zap](https://github.com/uber-go/zap) — Structured JSON logging
- **Tracing:** [Jaeger](https://www.jaegertracing.io/) + [OpenTelemetry](https://opentelemetry.io/) — Distributed request tracing
- **Log Aggregation:** [Loki](https://grafana.com/docs/loki/) — Log storage and querying
- **Visualization:** [Grafana](https://grafana.com/) — Dashboard for logs/metrics
- **Database:** [PostgreSQL 16 LTS](https://www.postgresql.org/) — Primary data store

## Project Structure

```
test_NF/
├── config/              # Configuration loader (env variables, defaults)
├── controller/          # HTTP request handlers and routing
├── service/             # Business logic and domain rules
├── repository/          # Database access layer (GORM queries)
├── model/               # Data structures (DTOs, database entities)
├── internal/
│   ├── database/        # Database initialization and pooling
│   ├── logging/         # Structured JSON logger with context awareness
│   ├── middleware/      # HTTP middleware (request tracing, correlation IDs)
│   └── observability/   # Jaeger tracing setup and initialization
├── documentation/       # API specs and documentation
├── main.go              # Application entry point
├── .env                 # Environment variables (local development)
├── docker-compose.yml   # Local infrastructure stack
├── loki-config.yaml     # Loki configuration
└── go.mod/go.sum        # Go dependencies
```

### Folder Objectives

| Folder | Purpose |
|--------|---------|
| **config** | Load and parse environment variables into typed config structs |
| **controller** | Handle HTTP requests, validate input, map status codes, delegate to services |
| **service** | Implement business logic, orchestrate repositories, handle validation |
| **repository** | Pure data access operations using GORM, no business logic |
| **model** | Define request/response DTOs and database entities |
| **internal/database** | Initialize GORM, configure connection pooling |
| **internal/logging** | Setup structured JSON logging with correlation IDs |
| **internal/middleware** | Add request tracing, generate/extract trace IDs |
| **internal/observability** | Initialize Jaeger for distributed tracing |

## Architecture

3-layer architecture with clear separation of concerns:

```
HTTP Request → Controller → Service → Repository → Database
```

- **Controller:** Routes and HTTP concerns
- **Service:** Business logic and validation
- **Repository:** Data access only
- **Dependency Injection:** Constructor-based, no global state

## Configuration

All settings via environment variables (see `.env`):

```bash
# Server
SERVER_PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=irrigationuser
DB_PASSWORD=irrigationpass
DB_NAME=irrigation_db

# Jaeger
JAEGER_AGENT_HOST=localhost
JAEGER_AGENT_PORT=6831
JAEGER_SAMPLER_TYPE=const
JAEGER_SAMPLER_PARAM=1

# Loki
LOKI_URL=http://localhost:3100
```

## Observability

### Structured Logging
- JSON format with ISO8601 timestamps
- Automatic correlation IDs (request_id, trace_id) via middleware
- Context-aware logging throughout request lifecycle

### Distributed Tracing
- Jaeger captures all requests for flow visualization
- Traces all requests by default (`JAEGER_SAMPLER_PARAM=1`)
- View traces in Jaeger UI at http://localhost:16686

## Development

### Add a New Endpoint

1. **Define model** → `model/entity.go`
2. **Create repository** → `repository/entity.go` (database queries)
3. **Implement service** → `service/entity.go` (business logic)
4. **Add controller** → `controller/entity.go` (HTTP handler)
5. **Register route** → `main.go` (wire everything)

See `.github/copilot-instructions.md` for detailed step-by-step guide.

### Testing

Create tests alongside source files:
- `*_test.go` — Use standard Go `testing` package
- Run with: `go test ./...`

## Common Commands

```bash
# Build
go build -o irrigation-api main.go

# Test
go test ./...

# Format
go fmt ./...

# Lint
golangci-lint run ./...

# View logs
docker-compose logs -f postgres
docker-compose logs -f loki

# Database access
psql -h localhost -U irrigationuser -d irrigation_db
```

## References

- [Copilot Instructions](.github/copilot-instructions.md) — Detailed architecture guide
- [Gin Docs](https://gin-gonic.com/)
- [GORM Docs](https://gorm.io/)
- [Jaeger Docs](https://www.jaegertracing.io/)

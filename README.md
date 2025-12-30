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
| **Promtail** | (internal) | Log shipper (pushes container logs to Loki) |
| **PostgreSQL** | localhost:5432 | Database |
| **Swagger UI** | http://localhost:8080/swagger/index.html | API docs viewer (navigate directly to /swagger/index.html) |
| **Swagger JSON** | http://localhost:8080/docs/swagger.json | OpenAPI spec |

To view logs in Grafana, go to http://localhost:3000 (admin/admin), add Loki as a datasource (http://loki:3100), and query logs.

### Database Seeding

**Schema Migration:**
- Database schema is created automatically on startup via GORM AutoMigrate
- Migrates `Farm`, `IrrigationSector`, and `IrrigationData` tables with optimized indexes
- Safe to run multiple times (AutoMigrate is idempotent)

**Seed Data:**
- Sample data available in [internal/seeds/irrigation_seed.json](internal/seeds/irrigation_seed.json)
- **1,680 irrigation records** spanning 3 years (2023-2025) with same date ranges for year-over-year comparisons
- 2 farms with 4 irrigation sectors each (8 sectors total)
- Irrigation season data: March 1 - October 31 (recorded on Tuesdays and Fridays each week)
- Seasonal variation: 20-35mm nominal amounts (higher in summer Jun-Aug)
- Efficiency profiles (real vs nominal water usage):
  - Excellent: North Grove (Farm 2) at 99%
  - Good: North Field (Farm 1) at 95%, East Orchard (Farm 2) at 92%
  - Average: South Field (Farm 1) at 88%, West Orchard (Farm 2) at 85%
  - Poor: East Pasture (Farm 1) at 75% (leaks), South Grove (Farm 2) at 80%
- Deterministic variation enables consistent year-over-year analysis
- CLI scripts provided in [internal/scripts/](internal/scripts/) folder:
  - `go run internal/scripts/seed.go` — Loads all seed data (idempotent)
  - `go run internal/scripts/cleanup.go` — Removes all seeded data
- Service layer provides reusable methods:
  - `FarmService.SeedFarms()` — Loads farms from JSON
  - `IrrigationSectorService.SeedSectors()` — Loads sectors
  - `IrrigationDataService.SeedData()` — Loads irrigation records
- Seeding is idempotent (uses GORM `Save()` to upsert)

**Performance:**
- See [DatabaseOptimization.md](documentation/DatabaseOptimization.md) for:
  - Composite index strategy for large datasets
  - SQL-level aggregation patterns
  - Query optimization techniques
  - Connection pool tuning


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
├── data/
│   └── seeds/           # Seed data JSON files for database initialization
├── internal/
│   ├── database/        # Database initialization, pooling, and AutoMigrate
│   ├── logging/         # Structured JSON logger with context awareness
│   ├── middleware/      # HTTP middleware (request tracing, correlation IDs)
│   ├── observability/   # Jaeger tracing setup and initialization
│   ├── scripts/         # CLI utilities (seed.go, cleanup.go)
│   └── seeds/           # JSON seed files for database initialization
├── documentation/       # Performance optimization guides
├── swagger/             # Swagger/OpenAPI specs and generated docs
├── main.go              # Application entry point
├── .env                 # Environment variables (local development)
├── docker-compose.yml   # Local infrastructure stack
├── loki-config.yaml     # Loki configuration
├── promtail-config.yaml # Promtail configuration (log shipping)
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
| **internal/seeds** | JSON seed files for deterministic database initialization |
| **documentation** | Performance optimization guides and best practices |
| **swagger** | Swagger/OpenAPI specs, generated documentation, and API stubs |
| **internal/database** | Initialize GORM, configure connection pooling, run AutoMigrate |
| **internal/logging** | Setup structured JSON logging with correlation IDs |
| **internal/middleware** | Add request tracing, generate/extract trace IDs |
| **internal/observability** | Initialize Jaeger for distributed tracing |
| **internal/scripts** | CLI utilities for database operations (seeding, cleanup) |

## Architecture

3-layer architecture with clear separation of concerns:

```
HTTP Request → Controller → Service → Repository → Database
```

- **Controller:** Routes and HTTP concerns
- **Service:** Business logic and validation
- **Repository:** Data access only
- **Dependency Injection:** Constructor-based, no global state

### Data Model

The system manages irrigation analytics across three core entities:

- **Farm** — Top-level agricultural properties
- **IrrigationSector** — Subdivisions of farms with irrigation capabilities (linked to farms)
- **IrrigationData** — Time-series irrigation event records tracking nominal vs. actual water usage (linked to farms and sectors)

All entities use GORM with optimized composite indexes for time-range queries. See [DatabaseOptimization.md](documentation/DatabaseOptimization.md) for indexing strategy and query patterns.

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
- Logs shipped to Loki via Promtail for centralized storage and querying

### Log Aggregation (Loki + Promtail)
- **Loki** collects and stores logs
- **Promtail** scrapes container logs from Docker and pushes them to Loki
- Queryable via Grafana at http://localhost:3000
- Query example in Grafana: `{job="docker", container_name=~"irrigation.*"}`

### Distributed Tracing
- Jaeger captures all requests for flow visualization
- Traces all requests by default (`JAEGER_SAMPLER_PARAM=1`)
- Includes database query spans via GORM OpenTelemetry plugin
- View traces in Jaeger UI at http://localhost:16686

## Development

### API Docs (Swagger)
- View UI: http://localhost:8080/swagger/index.html (no redirect from /swagger)
- View JSON: http://localhost:8080/docs/swagger.json
- Regenerate spec (JSON/YAML only):
	```bash
	go install github.com/swaggo/swag/cmd/swag@latest
	$(go env GOPATH)/bin/swag init --output ./swagger --dir ./ --outputTypes json,yaml
	```
	(Static files are served from `swagger/swagger.json` and `swagger/swagger.yaml`.)

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

# Seed database with sample data (idempotent)
go run internal/scripts/seed.go

# Clean up all seeded data
go run internal/scripts/cleanup.go

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

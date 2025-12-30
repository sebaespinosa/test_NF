# Irrigation Analytics API - Copilot Instructions

## Project Overview

This is a Go-based REST API for managing irrigation analytics within an agricultural platform. The service is built with production-ready patterns including structured logging, distributed tracing, database connection pooling, and health checks.

**Tech Stack:**
- **Framework:** Gin (lightweight HTTP framework)
- **ORM:** GORM (with PostgreSQL driver)
- **Observability:** OpenTelemetry + Jaeger (distributed tracing), Zap (structured logging), Loki (log aggregation)
- **Database:** PostgreSQL 16 (LTS)
- **Container:** Docker Compose (for local development)

---

## Architecture

### Folder Structure

```
test_NF/
├── config/                    # Configuration loader (environment variables)
├── controller/                # HTTP request handlers and routing
├── model/                     # Data structures and domain models
├── repository/                # Database access layer
├── service/                   # Business logic layer
├── documentation/             # API documentation and specs
├── internal/
│   ├── database/             # Database initialization and setup
│   ├── logging/              # Structured JSON logging with correlation IDs
│   ├── middleware/           # HTTP middleware (tracing, request IDs)
│   └── observability/        # Jaeger/OpenTelemetry tracing setup
├── main.go                   # Application entry point
├── go.mod                    # Go module definition
├── go.sum                    # Dependency lock file
├── .env                      # Environment variables (local dev)
├── docker-compose.yml        # Docker services for local development
├── loki-config.yaml         # Loki log aggregation configuration
└── .github/
    └── copilot-instructions.md  # This file
```

### Layered Architecture Pattern

The project follows a **3-layer architecture** with clear separation of concerns:

```
Controller (HTTP) → Service (Business Logic) → Repository (Data Access)
     ↓                    ↓                          ↓
  Gin Handlers      Business Rules            GORM/Database
```

**Layer Responsibilities:**

1. **Controller Layer** (`controller/`)
   - Handles incoming HTTP requests and responses
   - Maps HTTP status codes
   - Delegates to service layer for business logic
   - No database queries; no business logic

2. **Service Layer** (`service/`)
   - Implements all business logic
   - Orchestrates multiple repositories
   - Logs meaningful events with context
   - Handles errors and validation
   - Contains domain-specific rules

3. **Repository Layer** (`repository/`)
   - Pure data access operations
   - GORM query builders
   - No business logic; no HTTP concerns
   - Parameterized queries (automatic with GORM)

4. **Model Layer** (`model/`)
   - Request/response DTOs
   - Database entities (GORM models)
   - Data structures shared across layers

---

## Configuration Management

### Environment Variables (.env)

All configuration is loaded from environment variables via `config/config.go`:

- **Server:** `SERVER_PORT`, `ENV`
- **Database:** `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSL_MODE`, `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`
- **Jaeger:** `JAEGER_AGENT_HOST`, `JAEGER_AGENT_PORT`, `JAEGER_SAMPLER_TYPE`, `JAEGER_SAMPLER_PARAM`
- **Loki:** `LOKI_URL`
- **Service:** `SERVICE_NAME`, `SERVICE_VERSION`

**Default values** are provided for local development. Override via `.env` file or system environment.

### Configuration Loader Pattern

```go
// config.Load() returns a *Config struct with all nested configurations
cfg, err := config.Load()
// Access nested config: cfg.Database.Host, cfg.Server.Port, etc.
```

---

## Observability Stack

### Structured Logging (Zap + JSON)

All logs are **JSON structured** for compatibility with Loki and APM tools:

```go
logger.Info("operation completed",
    zap.String("user_id", userID),
    zap.Int("records", count),
    zap.Duration("elapsed", elapsed),
)
```

**Log Format:** Standard JSON with ISO 8601 timestamps.

**Correlation IDs:** Automatically injected via middleware:
- `request_id` — Unique per request (UUID or from `X-Request-ID` header)
- `trace_id` — Spans multiple requests (UUID or from `X-Trace-ID` header)

**Context-Aware Logging:**
```go
ctxLogger := logger.WithContext(ctx)  // Extracts correlation IDs from context
ctxLogger.Info("processing", zap.String("operation", "sync"))
```

### Distributed Tracing (Jaeger + OpenTelemetry)

Traces are exported to **Jaeger** for request flow visualization:

- **Sampling:** Default is `JAEGER_SAMPLER_TYPE=const` with `JAEGER_SAMPLER_PARAM=1` (trace all; adjust for production)
- **Agent Port:** UDP 6831 (default)
- **UI:** http://localhost:16686 (after `docker-compose up`)

**Instrumentation Strategy:**
- Middleware automatically creates request spans
- Service methods can create child spans for detailed tracing
- Context propagation via `context.Context`

### Log Aggregation (Loki)

Logs are sent to **Loki** for long-term storage and querying:

- **Push Model:** Applications push logs directly to Loki
- **Query:** Use Grafana or `logcli` tool
- **Retention:** Configured in `loki-config.yaml`

---

## Dependency Injection Pattern

This project uses **constructor injection** for all components:

```go
// Bad: Global state or global logger
logger = setupLogger()

// Good: Injected via constructor
type HealthService struct {
    repo   *repository.HealthRepository
    logger *logging.Logger
}

func NewHealthService(repo *repository.HealthRepository, logger *logging.Logger) *HealthService {
    return &HealthService{repo: repo, logger: logger}
}
```

**Benefits:**
- Explicit dependencies
- Testable (easy to mock)
- Clear initialization order in `main.go`
- No global state

---

## Error Handling

Follow Go's standard error handling practices:

```go
// Return errors with context
if err := db.Save(model).Error; err != nil {
    logger.Error("failed to save record", zap.Error(err))
    return nil, fmt.Errorf("save operation failed: %w", err)
}

// At HTTP boundary: map to status codes
if err := service.DoSomething(ctx); err != nil {
    ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
}
```

**Patterns:**
- Use `fmt.Errorf()` with `%w` for error wrapping
- Log errors with full context (user ID, request ID, operation)
- Status codes: 200 OK, 400 Bad Request, 404 Not Found, 500 Internal Server Error

---

## HTTP Request/Response Flow

### Request Tracing & Logging

All requests pass through the **TraceMiddleware**:

1. Extract or generate `X-Request-ID` and `X-Trace-ID` headers
2. Store in `gin.Context` for access in handlers
3. Inject into request context for logging
4. Log incoming request with method, path, correlation IDs
5. Continue to next handler
6. Log response with status code and correlation IDs

**Middleware Usage:**
```go
router.Use(middleware.TraceMiddleware(logger))
```

### Response Structure

**Success (2xx):**
```json
{
  "status": "healthy",
  "message": "service is running",
  "version": "0.0.1"
}
```

**Error (4xx/5xx):**
```json
{
  "error": "descriptive error message"
}
```

---

## Database Layer (GORM + PostgreSQL)

### Connection Setup

GORM is initialized with **PostgreSQL connection pooling**:

```go
db, err := database.Initialize(&cfg.Database)
// Connection pool configured:
// - MaxOpenConns: 25
// - MaxIdleConns: 5
// - ConnMaxLifetime: 5 minutes
```

### Model Pattern

GORM models use standard Go structs with tags:

```go
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Email     string    `gorm:"uniqueIndex"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### Repository Pattern

All database access goes through repositories:

```go
type UserRepository struct {
    db *gorm.DB
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
    var user User
    if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
        return nil, err
    }
    return &user, nil
}
```

**Key Points:**
- Always use `WithContext(ctx)` for proper timeout handling
- Use named parameters (`?`) for safety
- Return domain models, not GORM models

---

## Local Development Setup

### Prerequisites

- Go 1.22+ (check with `go version`)
- Docker and Docker Compose
- PostgreSQL client tools (optional: `psql`)

### Start Services

```bash
# Start all services (PostgreSQL, Jaeger, Loki, Grafana)
docker-compose up -d

# Verify services
docker-compose ps

# View logs
docker-compose logs -f postgres
docker-compose logs -f loki
```

### Run Application

```bash
# Install dependencies
go mod download

# Run with hot reload (requires entr or similar)
go run main.go

# Or with air (Go file watcher)
air
```

### Access Services

- **API:** http://localhost:8080
- **Health:** http://localhost:8080/health
- **Jaeger UI:** http://localhost:16686
- **Grafana:** http://localhost:3000 (admin/admin)
- **Loki:** http://localhost:3100

### Database Access

```bash
# Connect with psql
psql -h localhost -U irrigationuser -d irrigation_db

# Inside psql:
\dt              # List tables
\d table_name    # Describe table
```

---

## Adding New Features

### Step-by-Step: Add a New Endpoint

**1. Define Model** (`model/user.go`)
```go
type User struct {
    ID    uint   `json:"id" gorm:"primaryKey"`
    Email string `json:"email" gorm:"uniqueIndex"`
    Name  string `json:"name"`
}
```

**2. Create Repository** (`repository/user.go`)
```go
func (r *UserRepository) FindByID(ctx context.Context, id uint) (*User, error) {
    var user User
    if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
        return nil, err
    }
    return &user, nil
}
```

**3. Implement Service** (`service/user.go`)
```go
func (s *UserService) GetUser(ctx context.Context, id uint) (*User, error) {
    s.logger.WithContext(ctx).Info("fetching user", zap.Uint("user_id", id))
    return s.repo.FindByID(ctx, id)
}
```

**4. Create Controller** (`controller/user.go`)
```go
func (c *UserController) GetUser(ctx *gin.Context) {
    id := ctx.Param("id")
    user, err := c.service.GetUser(ctx.Request.Context(), uint(id))
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
    }
    ctx.JSON(http.StatusOK, user)
}
```

**Swagger Annotations:** Every new endpoint must include swag annotations on the controller handler with at least:
- `@Summary` and `@Description`
- `@Tags <feature>`
- `@Accept`/`@Produce` where relevant (commonly `json`)
- Request/response schemas via `@Param` and `@Success`/`@Failure` using model types
- `@Router /path [method]`

**Regenerating Swagger Spec:**
```
go install github.com/swaggo/swag/cmd/swag@latest
$(go env GOPATH)/bin/swag init --output ./documentation --dir ./ --outputTypes json,yaml
```
Static docs are served from `documentation/swagger.json` via `/swagger/doc.json` and the UI at `/swagger/index.html`.

**5. Register Route** (`main.go`)
```go
router.GET("/users/:id", userController.GetUser)
```

### Testing

When requested, create tests in:
- `repository/*_test.go` — Database layer tests
- `service/*_test.go` — Business logic unit tests
- `controller/*_test.go` — HTTP handler tests (integration)

Test files should use standard Go `testing` package.

---

## Code Standards & Best Practices

### Naming Conventions

- **Packages:** lowercase, single word (`repository`, `service`)
- **Types/Functions:** PascalCase (`GetUser`, `UserRepository`)
- **Variables/Constants:** camelCase (`userName`, `maxRetries`)
- **Constants:** UPPER_SNAKE_CASE for unexported package constants
- **File names (controllers/services/repositories/models):** use `<feature>_<layer>.go` (e.g., `health_controller.go`, `health_service.go`, `health_repository.go`, `health_model.go`) to avoid filename collisions across layers

### Error Handling

- Always wrap errors with context: `fmt.Errorf("operation: %w", err)`
- Log errors with all relevant context
- Return early on errors (fail-fast pattern)
- Don't swallow errors unless explicitly handled

### Logging

- Use structured logging (key-value pairs)
- Log at appropriate levels: Info, Warn, Error, Fatal
- Always include correlation IDs in logs
- Avoid logging sensitive data (passwords, tokens)

### Comments

- Exported functions must have doc comments
- Explain "why", not "what" (code shows what)
- Keep comments updated with code changes

---

## Deployment Checklist

- [ ] Update `SERVICE_VERSION` in `.env` / Dockerfile
- [ ] Review and test error scenarios
- [ ] Configure production-grade logging (adjust Jaeger sampler)
- [ ] Set up TLS for database connections (`DB_SSL_MODE=require`)
- [ ] Configure connection pool for production load
- [ ] Set up log retention policies in Loki
- [ ] Test graceful shutdown (SIGTERM handling)
- [ ] Review and update documentation

---

## Troubleshooting

**Database connection refused:**
- Ensure PostgreSQL is running: `docker-compose ps postgres`
- Check `.env` database credentials match `docker-compose.yml`
- Verify `DB_HOST=localhost` (not 127.0.0.1 if using Docker network)

**Jaeger not receiving traces:**
- Verify Jaeger is running: `docker-compose ps jaeger`
- Check `JAEGER_AGENT_HOST` and `JAEGER_AGENT_PORT` in `.env`
- Ensure sampler is not disabled: `JAEGER_SAMPLER_PARAM > 0`

**Logs not appearing in Loki:**
- Logs are pushed locally; Loki integration requires Promtail or Logstash
- For now, view logs via application output and JSON-structured format
- Configure Promtail in future for push-based log collection

---

## References

- [Gin Documentation](https://gin-gonic.com/)
- [GORM Guide](https://gorm.io/)
- [Zap Logger](https://github.com/uber-go/zap)
- [Jaeger Tracing](https://www.jaegertracing.io/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [PostgreSQL Docker Image](https://hub.docker.com/_/postgres)
- [Loki Documentation](https://grafana.com/docs/loki/)

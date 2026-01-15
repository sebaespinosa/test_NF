# Backend + Infrastructure Engineer Coding Assignment

# Backend & Infrastructure Engineer - Coding Assignment

## Overview

This assignment is designed to evaluate your skills in backend development, database optimization, and infrastructure engineering. You'll build a REST API service from scratch that handles irrigation analytics for a farming data platform.

**Time Limit:** 2-3 days  
**Submission:** Please provide a link to your repository (GitHub, GitLab, etc.) with your implementation

## Context

You'll be building a service that manages irrigation analytics for an agricultural platform. The system needs to:

1. Store irrigation data (irrigation events with nominal and real amounts)
2. Calculate irrigation efficiency metrics
3. Generate time-series analytics
4. Provide aggregated statistics for dashboards

## **Assignment: Irrigation Analytics Service**

### **Background**

Build a REST API service that provides analytics and monitoring for irrigation systems. The service should:

1. Store irrigation data efficiently
2. Calculate irrigation efficiency metrics
3. Generate time-series analytics with different aggregation levels
4. Provide aggregated statistics for dashboards
5. Handle large datasets efficiently

### **Requirements**

Implement the following features:

**1. Irrigation Analytics Endpoint**

Create a new REST API endpoint that provides irrigation analytics for a given farm or irrigation sector. The endpoint must include year-over-year comparisons to help identify anomalies and trends.

**Endpoint:** `GET /v1/farms/{farm_id}/irrigation/analytics`

**Query Parameters:**

- `start_date` (optional): ISO 8601 date string
- `end_date` (optional): ISO 8601 date string
- `sector_id` (optional): Filter by specific irrigation sector
- `aggregation` (optional): `daily`, `weekly`, `monthly` (default: `daily`)

**Response Format:**

```json
{
  "farm_id": 123,
  "period": {
    "start": "2024-01-01T00:00:00Z",
    "end": "2024-01-31T23:59:59Z"
  },
  "aggregation": "daily",
  "metrics": {
    "total_irrigation_volume_mm": 450.5,
    "total_irrigation_events": 120,
    "average_efficiency": 0.85,
    "efficiency_range": {
      "min": 0.72,
      "max": 0.98
    },
    "same_period_-1": {
      "total_irrigation_volume_mm": 420.3,
      "total_irrigation_events": 115,
      "average_efficiency": 0.82,
      "efficiency_range": {
        "min": 0.70,
        "max": 0.95
      }
    },
    "same_period_-2": {
      "total_irrigation_volume_mm": 480.1,
      "total_irrigation_events": 125,
      "average_efficiency": 0.88,
      "efficiency_range": {
        "min": 0.75,
        "max": 0.99
      }
    },
    "period_comparison": {
      "vs_same_period_-1": {
        "volume_change_percent": 7.2,
        "events_change_percent": 4.3,
        "efficiency_change_percent": 3.7
      },
      "vs_same_period_-2": {
        "volume_change_percent": -6.2,
        "events_change_percent": -4.0,
        "efficiency_change_percent": -3.4
      }
    }
  },
  "time_series": [
    {
      "date": "2024-01-01",
      "nominal_amount_mm": 12.5,
      "real_amount_mm": 10.8,
      "efficiency": 0.864,
      "event_count": 3
    }
    // ... more entries
  ],
  "sector_breakdown": [
    {
      "sector_id": 1,
      "sector_name": "Sector A",
      "total_volume_mm": 150.2,
      "average_efficiency": 0.88
    }
  ]
}
```

**Efficiency Calculation:**

- `efficiency = real_amount / nominal_amount`
- Handle edge cases (zero nominal amount, null values)

**Year-over-Year Comparison Requirements:**

- Calculate metrics for the same date range in the previous year (`same_period_-1`)
- Calculate metrics for the same date range two years ago (`same_period_-2`)
- For example, if the requested period is `2024-01-01` to `2024-01-31`:
    - `same_period_-1` should compare with `2023-01-01` to `2023-01-31`
    - `same_period_-2` should compare with `2022-01-01` to `2022-01-31`
- Calculate percentage changes for key metrics:
    - `volume_change_percent = ((current - previous) / previous) * 100`
    - `events_change_percent = ((current - previous) / previous) * 100`
    - `efficiency_change_percent = ((current - previous) / previous) * 100`
- Handle cases where previous period data doesn't exist (return null or omit the comparison)
- Handle division by zero when calculating percentage changes
- The comparison should use the same aggregation level (daily, weekly, monthly) as the current period

**2. Database Optimization**

The `irrigation_data` table can grow very large (millions of rows). Optimize queries for the analytics endpoint:

**Requirements:**

- Ensure queries perform well even with large datasets (consider indexes)
- Use efficient aggregation queries (avoid N+1 problems)
- Consider database-level aggregations vs application-level
- Add appropriate database indexes if needed
- Document your optimization decisions

**Constraints:**

- Consider both read and write performance
- Think about query patterns (time-range queries, grouping by sector, etc.)

**3. Monitoring & Observability**

Add observability to the new endpoint:

**Requirements:**

- Add structured logging for the analytics endpoint (include timing, query parameters, result counts)
- Add basic metrics (request count, latency, error rate) - you can use a simple in-memory counter or integrate with a metrics library
- Handle errors gracefully with appropriate HTTP status codes
- Add request validation with clear error messages

**4. Code Quality & Architecture**

**Requirements:**

- Follow existing code patterns and architecture (controller → service → repository)
- Add unit tests for the analytics calculation logic
- Add integration tests for the endpoint (at least happy path)
- Write clear, self-documenting code
- Add appropriate comments for complex logic
- Consider edge cases (empty data, invalid date ranges, etc.)

**5. Documentation**

**Requirements:**

- Update or create API documentation (Swagger/OpenAPI if possible)
- Add a brief README section explaining the new feature
- Document any database changes (migrations, indexes)

## **Technical Requirements**

### **Technology Stack**

1. **Language:** Go 1.21+
2. **Framework:** Gin (for HTTP server)
3. **Database:** PostgreSQL
4. **ORM:** GORM
5. **Code Style:** Follow Go conventions and best practices

### **Database Schema**

You need to create the following tables. Use GORM models and migrations:

**Farms Table:**

```go
type Farm struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"not null"`
}

```

**Irrigation Sectors Table:**

```go
type IrrigationSector struct {
    ID     uint   `gorm:"primaryKey"`
    FarmID uint   `gorm:"not null;index"`
    Name   string `gorm:"not null"`
    Farm   Farm   `gorm:"foreignKey:FarmID"`
}

```

**Irrigation Data Table:**

```go
type IrrigationData struct {
    ID                 uint      `gorm:"primaryKey"`
    FarmID             uint      `gorm:"not null;index"`
    IrrigationSectorID uint      `gorm:"not null;index"`
    StartTime          time.Time `gorm:"not null;index"`
    EndTime            time.Time `gorm:"not null"`
    NominalAmount      float32   `gorm:"type:numeric(10,2)"` // in mm
    RealAmount         float32   `gorm:"type:numeric(10,2)"` // in mm
    CreatedAt          time.Time
    UpdatedAt          time.Time
    Farm               Farm               `gorm:"foreignKey:FarmID"`
    IrrigationSector   IrrigationSector   `gorm:"foreignKey:IrrigationSectorID"`
}

```

**Important:** Consider appropriate indexes for query performance (especially for time-range queries and filtering by farm/sector).

### **Reverse Proxy & HTTPS/TLS Configuration**

The service **must** be served over HTTPS using TLS certificates, but TLS termination should be handled by a **reverse proxy** component, not in the Go application itself.

**Requirements:**

- The Go application should serve HTTP (not HTTPS) on an internal port (e.g., 8080)
- Set up a reverse proxy (nginx or similar) to handle TLS termination
- The reverse proxy should:
    - Terminate TLS/SSL connections
    - Forward requests to the Go application over HTTP
    - Serve HTTPS on port 443 (or 8443 for development)
    - Handle certificate loading and management
- Support both development (self-signed certificates) and production scenarios
- Document how to generate/obtain certificates for local development

**Implementation Notes:**

- This is a common production pattern that separates concerns (app handles business logic, proxy handles TLS)
- The Go application remains simple and doesn't need to manage certificates
- Certificate rotation and management is handled at the proxy level
- Consider using environment variables or configuration files for proxy settings

### **Project Structure**

Organize your code following a clean architecture pattern:

```
your-project/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── controller/          # HTTP handlers
│   ├── service/             # Business logic
│   ├── repository/          # Data access layer
│   └── model/              # Data models
├── migrations/             # Database migrations (optional)
├── pkg/                    # Shared utilities (optional)
├── docker-compose.yml      # Local development setup
├── Dockerfile              # Container definition
├── go.mod
├── go.sum
└── README.md

```

## **Evaluation Criteria**

Your submission will be evaluated on:

1. **Functionality (40%)**
    - Does the feature work correctly?
    - Are edge cases handled?
    - Is the API design clean and intuitive?
2. **Performance & Scalability (25%)**
    - Are database queries optimized?
    - Will the solution scale with large datasets?
    - Are there any obvious performance bottlenecks?
3. **Code Quality (20%)**
    - Code organization and architecture
    - Test coverage
    - Error handling
    - Code readability
4. **Infrastructure & Observability (10%)**
    - Reverse proxy setup and TLS termination
    - Logging and monitoring
    - Error handling
    - Operational considerations
    - Security best practices
    - Infrastructure architecture (separation of concerns)
5. **Documentation (5%)**
    - Code documentation
    - API documentation
    - README updates

---

## **Additional Requirements**

### **Data Seeding**

Create a way to seed test data (script, endpoint, or migration). Include:

- At least 2 farms
- At least 3 irrigation sectors per farm
- At least 1000 irrigation data records spanning **multiple years** (minimum 3 years of data)
- Data should cover the same date ranges across different years to enable year-over-year comparisons
- Vary the efficiency values (some realistic, some edge cases)
- Include data that will allow testing of year-over-year comparison features

### **Docker Setup**

Provide a `docker-compose.yml` that:

- Sets up PostgreSQL
- Runs your application (serving HTTP on internal port)
- Includes a reverse proxy (nginx or similar) for TLS termination
- Makes it easy to test the service locally
- Handles TLS certificates (either via volume mounts or generation script)
- The reverse proxy should expose HTTPS on port 8443 (or 443) and forward to the app

---

## **Deliverables**

1. **Code Implementation**
    - All source code changes
    - Database migrations (if any)
    - Tests
2. **Documentation**
    - API documentation
    - Brief explanation of design decisions
    - Any database optimization notes
    - Reverse proxy configuration documentation
    - Instructions for generating/setting up TLS certificates for local development
3. **Brief Summary** (in README or separate file)
    - What you implemented
    - Key design decisions
    - Any trade-offs or limitations
    - What you would improve given more time

---

## **Tips**

- **Start simple:** Get a basic endpoint working first, then optimize
- **Database first:** Design your schema and indexes carefully - this is critical for performance
- **Test incrementally:** Test each layer (repository → service → controller) independently
- **Performance testing:** Use `EXPLAIN ANALYZE` in PostgreSQL to verify query performance
- **Realistic data:** Create seed data that represents real-world scenarios, including data spanning multiple years for year-over-year comparisons
- **Year-over-year logic:** When implementing comparisons, carefully handle date calculations (account for leap years, month boundaries, etc.)
- **Error handling:** Think about what can go wrong and handle it gracefully (missing historical data, division by zero in percentage calculations, etc.)
- **Don't over-engineer:** Focus on clean, maintainable code that solves the problem
- **Documentation:** Clear README and code comments help us understand your decisions

## **Example API Usage**

Once implemented, your service should handle requests over HTTPS:

```bash
# Get daily analytics for a farm (using -k to skip certificate verification for self-signed certs in dev)
curl -k "https://localhost:8443/v1/farms/1/irrigation/analytics?start_date=2024-01-01&end_date=2024-01-31&aggregation=daily"

# Get weekly analytics for a specific sector
curl -k "https://localhost:8443/v1/farms/1/irrigation/analytics?sector_id=5&aggregation=weekly"

# Get monthly analytics
curl -k "https://localhost:8443/v1/farms/1/irrigation/analytics?aggregation=monthly"

```

**Note:** The `-k` flag is only needed for self-signed certificates in development. In production with proper CA-signed certificates, this flag is not needed.

---

## **Questions?**

If you have any questions about the assignment, please reach out. We're here to help!

Good luck! 🚀
# Integration Testing Strategy - Irrigation Analytics API

This guide covers end-to-end and API integration tests for the irrigation analytics endpoint, built from the scenarios in ToTest.md. These tests validate real PostgreSQL behavior (DATE_TRUNC, casts, aggregation) that is not covered by unit tests.

## Goals
- Validate analytics calculations against a real PostgreSQL instance
- Exercise HTTP boundary (Gin routing, middleware, status codes)
- Verify YoY comparisons, sector breakdowns, pagination, and aggregation
- Catch regressions that mocks/SQLite cannot detect

## Environment & Prerequisites
- Go 1.25.5+
- Docker & Docker Compose
- jq (JSON assertions used by runner script)
- Base URL: http://localhost:8080 (adjust if different)
- Database: PostgreSQL (from docker-compose). Ensure it is seeded with the analytics dataset (same data used for manual validation in ToTest.md).

### Bring up services
```bash
docker-compose up -d postgres jaeger loki grafana promtail
```

### Seed the database
```bash
go run internal/scripts/seed.go
```

### Run the API
```bash
go mod download
go run main.go
```

## Tooling
- `curl` for HTTP requests
- `jq` for JSON assertions
- Optional: `internal/scripts/run_integration.sh` to batch-run the scenarios below
- Optional: `newman`/Postman collection or `k6` for load/perf (not required for base suite)

### Quick runner (recommended)
After the API is running locally and the DB is seeded:
```bash
bash internal/scripts/run_integration.sh
```
Env overrides: `BASE_URL` (default http://localhost:8080), `FARM_ID` (default 1), `SECTOR_ID` (default 1).

## Test Data Assumptions
- The database contains multi-year irrigation records (2023–2025) with multiple sectors per farm.
- Efficiency edge cases exist (nominal_amount=0, real_amount > nominal_amount, gaps in data).
- Pagination scenarios assume more than one page of time-series buckets.

## Core Scenarios (from ToTest.md)

| # | Purpose | Endpoint / Params | Expected |
|---|---------|-------------------|----------|
| 1 | Default query (last 90 days) | `GET /v1/farms/1/irrigation/analytics` | 200/206; aggregation=daily; page=1 limit=50; all top-level keys present |
| 2 | Weekly aggregation custom range | `start_date=2024-03-01&end_date=2024-03-31&aggregation=weekly` | 200; period start/end match; 5 buckets (ISO weeks); YoY complete |
| 3 | Monthly aggregation | `start_date=2024-01-01&end_date=2024-03-31&aggregation=monthly` | 200; 3 buckets (Jan/Feb/Mar) |
| 4 | Sector filter | `sector_id=1` | 200; sector_breakdown length=1; names/metrics match sector |
| 5 | Pagination page 1 | `limit=5&page=1` with March range | 200; 5 items; total_count > 5; total_pages correct |
| 6 | Pagination page 2 | `limit=5&page=2` | 200; items follow page 1 order |
| 7 | YoY complete (200) | `start_date=2024-03-01&end_date=2024-03-31` | 200; YoY periods populated; data_incomplete=false |
| 8 | YoY incomplete (206) | `start_date=2023-01-01&end_date=2023-01-31` | 206; missing YoY marked data_incomplete=true |
| 9 | Invalid date | `start_date=invalid` | 400 with error message |
|10 | Invalid aggregation | `aggregation=yearly` | 400 with allowed values listed |
|11 | Farm not found | `farm_id=99999` | 404 |
|12 | Efficiency calc check | `start_date=2024-03-01&end_date=2024-03-05` | 200; efficiency = real/nominal; null when nominal=0 |
|13 | Empty result set | `start_date=2020-01-01&end_date=2020-01-31` | 200; zero/empty metrics and arrays |
|14 | Percentage change math | `start_date=2024-03-01&end_date=2024-03-31` | 200; percentage calculations correct; handle previous=0 as null |

## Execution Examples
Use the commands from ToTest.md; a few key ones:

Default query:
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics" | jq '.{status:.status, agg:.aggregation, page:.time_series.pagination.page, limit:.time_series.pagination.limit}'
```

Weekly aggregation:
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-31&aggregation=weekly" | jq '.time_series.data | length'
```

Invalid aggregation:
```bash
curl -i "http://localhost:8080/v1/farms/1/irrigation/analytics?aggregation=yearly"
```

## Validation Checklist
- HTTP status matches expectations (200/206/400/404)
- `aggregation` echoes request (daily/weekly/monthly only)
- Pagination metadata: page, limit, total_count, total_pages
- YoY fields: `same_period_-1` and `same_period_-2` data_incomplete flags
- Metrics: totals, averages, min/max, and percentage changes
- Sector breakdown respects `sector_id` filter and names
- Empty ranges return zero/empty structures without errors

## Performance & Observability (optional)
- Measure response time for limit=all and long ranges; target <2s with seeded dataset
- Watch DB connections (pool not exhausted) and logs for errors
- Trace requests via Jaeger when observability stack is up

## Coverage Gaps / Next Steps
- Automate these scenarios in a dedicated integration test suite (Go `testing` with real DB or Postman/Newman collection)
- Add CI job that spins up PostgreSQL via docker-compose, seeds data, runs integration suite
- Add load/perf runs (k6) for limit=all and long date ranges

## References
- ToTest.md (source scenarios)
- README.md (local setup)
- UnitTesting.md (complementary unit test plan)

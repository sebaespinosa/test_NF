# Unit Testing Strategy

Scope: repository, service, and controller layers for the analytics endpoint. Integration tests stay in ToTest.md.

## Tooling
- Standard `testing` package
- Assertions: `github.com/stretchr/testify/{assert,require}`
- Mocks: `go.uber.org/mock/mockgen` (generate interfaces/fakes for repositories/services)
- HTTP tests: `net/http/httptest` + Gin router setup
- DB tests (repo): in-memory SQLite via `gorm.io/driver/sqlite` with AutoMigrate

## Layer Plans

### Repository (irrigation_data_repository.go)
- Use SQLite in-memory; AutoMigrate `Farm`, `IrrigationSector`, `IrrigationData`.
- Seed minimal fixtures (few rows) instead of full `internal/seeds/irrigation_seed.json` to keep tests fast.
- **⚠️ Important:** Analytics aggregation methods (`GetAnalyticsForFarmByDateRange`, `GetYoYComparison`, `GetSectorBreakdownForFarm`) use PostgreSQL-specific SQL (DATE_TRUNC, ::numeric, EXTRACT) which SQLite does not support. Unit tests focus on basic CRUD operations; analytics aggregations require integration tests with actual PostgreSQL or `sqlmock`.
- Cover basic cases: time-range filtering, ordering, record creation.
- Advanced aggregation scenarios (daily/weekly/monthly grouping, YoY comparisons, sector filtering) deferred to integration tests.

### Service (irrigation_analytics_service.go)
- Mock repository with gomock; assert orchestration:
  - Defaults (nil dates → last 90 days)
  - Error propagation when repo fails
  - Efficiency null-handling and percentage change math
  - YoY completeness flags drive status decisions
- Convert slices/maps to response DTOs correctly (time series, sector breakdown).

### Controller (analytics_controller.go)
- Use httptest with a real Gin engine + mocked service.
- Validate parsing/validation of query params (dates, aggregation, page/limit, sector_id).
- Verify status codes: 200 vs 206 (data_incomplete), 400 for invalid params, 404 passthrough, 500 on service error.
- Ensure `limit=all` → 10000 behavior and pagination defaults.

## Mock Generation
- Install tool (if needed): `go install go.uber.org/mock/mockgen@latest`
- Example: `mockgen -source=repository/irrigation_data_repository.go -destination=internal/mocks/irrigation_data_repository_mock_test.go -package=mocks`
- Keep mocks in `internal/mocks` and exclude from production builds.

## Test Data Guidance
- Prefer tiny inline fixtures in tests (3–10 rows) to keep executions <100ms.
- Reserve full seed file for integration/performance runs only.
- Use deterministic timestamps to make YoY comparisons stable.

## Running Tests
- All unit tests: `go test ./...`
- Targeted layers:
  - Repository: `go test ./repository -run TestIrrigationDataRepository`
  - Service: `go test ./service -run TestIrrigationAnalyticsService`
  - Controller: `go test ./controller -run TestAnalyticsController`

## Future CI Hooks
- Add unit tests to CI before integration/perf suites.
- Cache Go build/test modules for faster runs.
- Optional: add coverage flags `go test ./... -coverprofile=coverage.out`.

Assumptions:
- No CORS, Rate Limiting, Authentication
- Postgres without RLS
- No separate database for time-series (recommended)
- No commit or branching strategy
- No hardening patterns like retry, circuit-braker
- No migration strategy for the database
- No adapters/abstraction layers or patterns added for external componentes like Loki for logging or OpenTelemetry
- No api versioning or segmentation
- No caching strategy

Definitions:
- Environment files not excluded for simplicity
- Store logs with Loki
- Use Jaeger for APM and tracing with OpenTelemetry
- Filename with folder suffix for controller, service, repository, model folders
- Add Pagination to the analytics endpoint for performance
- Custom status codes (206 for example) to handle cases like empty responses for non-existing data on the time range

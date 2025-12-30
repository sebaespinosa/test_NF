Assumptions:
- No CORS, Rate Limiting, Authentication
- Postgres without RLS
- No separate database for time-series (recommended)
- No commit or branching strategy
- No hardening patterns like retry, circuit-braker
- No migration strategy for the database

Definitions:
- Loki
- Jaeger with OpenTelemetry
- 
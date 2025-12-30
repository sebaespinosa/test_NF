# Database Optimization Guide

## Overview

This document outlines the database optimization strategies, indexing patterns, and query design decisions implemented in the Irrigation Analytics API to handle large datasets efficiently.

## Data Model

The irrigation analytics system uses three main entities:

1. **Farm** - Top-level entity representing agricultural properties
2. **IrrigationSector** - Subdivisions of farms with irrigation capabilities
3. **IrrigationData** - Time-series irrigation event records (high-volume)

### Expected Scale

- **Farms**: Low cardinality (hundreds to thousands)
- **IrrigationSectors**: Medium cardinality (thousands to tens of thousands)
- **IrrigationData**: High cardinality (millions to billions of rows)

The `irrigation_data` table is the primary performance concern due to its expected size and query patterns.

---

## Indexing Strategy

### 1. Composite Indexes for Time-Range Queries

**Problem**: Time-range analytics queries filtering by farm or sector are the most common access pattern.

**Solution**: Composite indexes that combine entity IDs with timestamps:

```go
// In IrrigationData model
`gorm:"index:idx_irrigation_farm_time,priority:1"`    // farm_id (first column)
`gorm:"index:idx_irrigation_farm_time,priority:2"`    // start_time (second column)
```

**Rationale**:
- PostgreSQL can use composite indexes efficiently for queries like:
  ```sql
  WHERE farm_id = ? AND start_time >= ? AND start_time <= ?
  ```
- The index stores data in sorted order by `(farm_id, start_time)`, allowing range scans
- Eliminates need for table scans when filtering by farm and time

**Query Coverage**:
- ✅ Filter by farm + time range
- ✅ Filter by farm only (uses leftmost prefix)
- ❌ Filter by time only (requires separate index)

### 2. Sector-Based Composite Index

```go
`gorm:"index:idx_irrigation_sector_time,priority:1"`  // irrigation_sector_id
`gorm:"index:idx_irrigation_sector_time,priority:2"`  // start_time
```

**Rationale**:
- Supports sector-level analytics independently of farm-level queries
- Enables efficient drill-down from farm → sectors → time-range data

### 3. Standalone Time Index

```go
`gorm:"index:idx_irrigation_time"`  // start_time only
```

**Rationale**:
- Supports global time-range queries without entity filters
- Used for system-wide analytics and reporting
- Relatively low overhead since time fields are small

### 4. Foreign Key Indexes

```go
`gorm:"index:idx_irrigation_farm"`      // farm_id
`gorm:"index:idx_irrigation_sector"`    // irrigation_sector_id
```

**Rationale**:
- Optimizes JOIN operations for aggregation queries
- PostgreSQL automatically creates indexes for foreign keys, but explicit definition ensures coverage
- Supports cascading deletes efficiently

---

## Query Optimization Patterns

### 1. SQL-Level Aggregations

**Problem**: Fetching all records and aggregating in application memory causes:
- High network transfer overhead
- Memory exhaustion with large datasets
- Slow response times

**Solution**: Push aggregations to PostgreSQL with GROUP BY:

```go
func (r *IrrigationDataRepository) AggregateByFarm(ctx context.Context, startTime, endTime time.Time) ([]FarmAggregation, error) {
    return r.db.WithContext(ctx).
        Table("irrigation_data").
        Select(`
            irrigation_data.farm_id,
            farms.name as farm_name,
            COUNT(*) as total_events,
            SUM(irrigation_data.nominal_amount) as total_nominal_amount,
            AVG(irrigation_data.nominal_amount) as avg_nominal_amount
        `).
        Joins("JOIN farms ON farms.id = irrigation_data.farm_id").
        Where("irrigation_data.start_time >= ? AND irrigation_data.start_time <= ?", startTime, endTime).
        Group("irrigation_data.farm_id, farms.name").
        Scan(&results).Error
}
```

**Benefits**:
- Database performs aggregation using efficient columnar operations
- Only aggregated results transferred over network (minimal data)
- Leverages composite indexes for WHERE clause filtering

### 2. Avoiding N+1 Queries

**Problem**: Loading irrigation data with relations causes N+1 queries:
```go
// BAD: Generates 1 + N queries
data := repo.FindAll()  // 1 query
for _, d := range data {
    d.Farm.Name  // N queries (one per row)
}
```

**Solution**: Use GORM's `Preload` or JOINs for aggregations:
```go
// GOOD: Single query with JOIN
data := repo.FindByID(ctx, id)  // Preload("Farm").Preload("IrrigationSector")
```

**For Analytics**: Use JOINs in aggregation queries (see example above) instead of preloading full objects.

### 3. Time-Range Query Optimization

**Best Practices**:
- Always use indexed columns in WHERE clauses (`farm_id`, `start_time`)
- Use `>=` and `<=` for inclusive ranges (supports index range scans)
- Order by indexed columns when possible (`ORDER BY start_time`)

**Example Query Plan** (PostgreSQL EXPLAIN):
```
Index Scan using idx_irrigation_farm_time on irrigation_data
  Index Cond: (farm_id = 1 AND start_time >= '2024-01-01' AND start_time <= '2024-01-31')
```
✅ Uses composite index (efficient range scan)

---

## Read vs. Write Performance Trade-offs

### Write Performance Impact

**Indexes slow down writes**:
- Each index must be updated on INSERT/UPDATE
- 5 indexes on `irrigation_data` = 5x write overhead

**Mitigation Strategies**:
1. **Batch Inserts**: Use GORM's `CreateInBatches()` for bulk loading
2. **Async Writes**: Consider message queues for high-throughput ingestion
3. **Partitioning**: For extreme scale, partition `irrigation_data` by time (e.g., monthly)

**Current Trade-off**: Prioritize read performance for analytics workload; accept modest write overhead.

### Index Selectivity

**High-selectivity indexes** (few duplicates) are most effective:
- ✅ `start_time`: High selectivity (unique timestamps)
- ✅ `(farm_id, start_time)`: Very high selectivity
- ⚠️ `farm_id` alone: Low selectivity (many rows per farm)

**Recommendation**: Composite indexes provide better selectivity for our query patterns.

---

## Query Patterns and Index Usage

| Query Pattern | Index Used | Performance |
|---------------|------------|-------------|
| Filter by farm + time range | `idx_irrigation_farm_time` | ⚡ Excellent |
| Filter by sector + time range | `idx_irrigation_sector_time` | ⚡ Excellent |
| Filter by time only | `idx_irrigation_time` | ✅ Good |
| Aggregate by farm (with time filter) | `idx_irrigation_farm_time` + JOIN | ⚡ Excellent |
| Aggregate by sector (with time filter) | `idx_irrigation_sector_time` + JOIN | ⚡ Excellent |
| Full table scan | None (sequential scan) | ❌ Poor (avoid) |

---

## Connection Pooling

**Configuration** (from `internal/database/db.go`):
```go
sqlDB.SetMaxOpenConns(25)       // Max concurrent connections
sqlDB.SetMaxIdleConns(5)        // Idle connections to keep alive
sqlDB.SetConnMaxLifetime(5min)  // Connection reuse lifetime
```

**Rationale**:
- **MaxOpenConns=25**: Balances concurrency with PostgreSQL connection limits
- **MaxIdleConns=5**: Reduces connection setup overhead for frequent queries
- **ConnMaxLifetime=5min**: Prevents stale connections and supports PostgreSQL connection recycling

**Production Tuning**:
- Monitor connection pool saturation (`db.Stats()`)
- Increase `MaxOpenConns` if pool exhaustion occurs under load
- Keep `MaxIdleConns` low to reduce memory overhead

---

## Schema Design Decisions

### Foreign Key Constraints

All relationships use `ON DELETE CASCADE`:
```go
`gorm:"foreignKey:FarmID;constraint:OnDelete:CASCADE"`
```

**Rationale**:
- Ensures referential integrity
- Simplifies cleanup operations (deleting farm removes sectors and data)
- Indexed foreign keys optimize JOIN performance

### Numeric Precision

```go
NominalAmount float32 `gorm:"type:numeric(10,2)"`
```

**Rationale**:
- PostgreSQL `numeric(10,2)` ensures precision for irrigation amounts (up to 99,999,999.99 mm)
- Avoids floating-point rounding errors in aggregations
- Sufficient range for real-world irrigation measurements

---

## Future Optimization Considerations

### Partitioning (for extreme scale)

If `irrigation_data` exceeds 100M rows, consider **time-based partitioning**:
```sql
CREATE TABLE irrigation_data_2024_01 PARTITION OF irrigation_data
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

**Benefits**:
- Queries filtered by time only scan relevant partitions
- Old partitions can be archived or dropped
- Maintenance operations (VACUUM, REINDEX) are faster

### Materialized Views

For frequently accessed aggregations (e.g., monthly summaries):
```sql
CREATE MATERIALIZED VIEW monthly_farm_summary AS
SELECT farm_id, DATE_TRUNC('month', start_time) AS month, SUM(real_amount)
FROM irrigation_data
GROUP BY farm_id, month;
```

**Benefits**:
- Pre-computed results for common queries
- Refresh on schedule (e.g., daily) instead of real-time computation

### Read Replicas

For high read traffic:
- Primary database handles writes
- Read replicas serve analytics queries
- Reduces load on primary database

---

## Monitoring and Profiling

### EXPLAIN ANALYZE

Test query performance with PostgreSQL's `EXPLAIN ANALYZE`:
```sql
EXPLAIN ANALYZE
SELECT farm_id, COUNT(*)
FROM irrigation_data
WHERE start_time >= '2024-01-01' AND start_time <= '2024-01-31'
GROUP BY farm_id;
```

**Look for**:
- ✅ "Index Scan" or "Bitmap Index Scan"
- ❌ "Seq Scan" (indicates missing index)

### Slow Query Log

Enable PostgreSQL slow query logging:
```sql
ALTER DATABASE irrigation_db SET log_min_duration_statement = 1000;  -- Log queries > 1s
```

### Application-Level Metrics

Monitor with OpenTelemetry spans:
- Query execution time
- Number of rows returned
- Connection pool stats

---

## Best Practices Summary

1. ✅ **Always filter by indexed columns** (`farm_id`, `irrigation_sector_id`, `start_time`)
2. ✅ **Use SQL-level aggregations** (COUNT, SUM, AVG) instead of application-level loops
3. ✅ **Avoid N+1 queries** with Preload or JOINs
4. ✅ **Use composite indexes** for multi-column filters
5. ✅ **Batch writes** when inserting large datasets
6. ✅ **Monitor query plans** with EXPLAIN to verify index usage
7. ✅ **Keep connection pool tuned** to application load
8. ⚠️ **Consider partitioning** for tables exceeding 100M rows
9. ⚠️ **Limit result sets** with pagination for user-facing queries
10. ❌ **Never SELECT * on large tables** without WHERE clauses

---

## References

- [PostgreSQL Indexing](https://www.postgresql.org/docs/current/indexes.html)
- [GORM Performance](https://gorm.io/docs/performance.html)
- [PostgreSQL Query Performance](https://www.postgresql.org/docs/current/using-explain.html)

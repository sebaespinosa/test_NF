# Irrigation Analytics Endpoint Guide

## Overview

The irrigation analytics endpoint provides comprehensive irrigation metrics with year-over-year comparisons, time-series aggregation, and sector-level breakdown. The endpoint analyzes water usage efficiency across multiple time periods to identify trends and anomalies.

## Endpoint

```
GET /v1/farms/:farm_id/irrigation/analytics
```

### Parameters

#### Path Parameters
- **farm_id** (required): Numeric farm identifier

#### Query Parameters
- **start_date** (optional): Analysis period start date in `YYYY-MM-DD` format
  - Default: 90 days before today
  - Example: `2024-01-01`
  - Time interpreted as 00:00:00 UTC

- **end_date** (optional): Analysis period end date in `YYYY-MM-DD` format
  - Default: Today
  - Example: `2024-01-31`
  - Time interpreted as 23:59:59 UTC

- **sector_id** (optional): Filter results to specific irrigation sector ID
  - Default: All sectors in farm
  - Example: `5`
  - When provided, only this sector appears in sector_breakdown array for better performance

- **aggregation** (optional): Time-series aggregation granularity
  - Valid values: `daily`, `weekly`, `monthly`
  - Default: `daily`
  - Determines how data is grouped in time_series array
  - Uses PostgreSQL `DATE_TRUNC` function for efficiency

- **page** (optional): Page number for pagination (1-indexed)
  - Default: `1`
  - Example: `2`

- **limit** (optional): Results per page
  - Default: `50`
  - Maximum: `1000`
  - Special value: `all` returns all results (may exceed timeout on large datasets >100k records)
  - Example: `50`

## Response Format

### Success Response (HTTP 200)

Returned when current year has complete data and at least one previous year (1-2 years ago) also has data available.

```json
{
  "farm_id": 1,
  "farm_name": "Green Valley Farm",
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
    }
  },
  "same_period_-1": {
    "total_irrigation_volume_mm": 420.3,
    "total_irrigation_events": 115,
    "average_efficiency": 0.82,
    "efficiency_range": {
      "min": 0.70,
      "max": 0.95
    },
    "data_incomplete": false
  },
  "same_period_-2": {
    "total_irrigation_volume_mm": 480.1,
    "total_irrigation_events": 125,
    "average_efficiency": 0.88,
    "efficiency_range": {
      "min": 0.75,
      "max": 0.99
    },
    "data_incomplete": false
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
  },
  "time_series": {
    "data": [
      {
        "date": "2024-01-01",
        "nominal_amount_mm": 12.5,
        "real_amount_mm": 10.8,
        "efficiency": 0.864,
        "event_count": 3
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 50,
      "total_count": 31,
      "total_pages": 1
    }
  },
  "sector_breakdown": [
    {
      "sector_id": 1,
      "sector_name": "North Field",
      "total_volume_mm": 150.2,
      "average_efficiency": 0.88
    },
    {
      "sector_id": 2,
      "sector_name": "South Field",
      "total_volume_mm": 120.5,
      "average_efficiency": 0.82
    }
  ]
}
```

### Partial Content Response (HTTP 206)

Returned when current period has data but previous year(s) are missing or incomplete. Indicates year-over-year comparisons may not be fully available.

The response structure is identical to HTTP 200, but:
- `data_incomplete` field in `same_period_-1` or `same_period_-2` is `true`
- `note` field explains why data is missing (e.g., "No data available for previous year (2023)")
- Corresponding comparison percentages in `period_comparison` may be `null`

### Error Responses

#### 400 Bad Request
- Invalid query parameters
- Malformed date format (must be `YYYY-MM-DD`)
- Invalid aggregation type
- Invalid pagination parameters

```json
{
  "error": "invalid start_date format; use YYYY-MM-DD"
}
```

#### 404 Not Found
- Farm ID does not exist

```json
{
  "error": "farm not found"
}
```

#### 500 Internal Server Error
- Database query failure
- Server processing error

## Data Definitions

### Metrics

All metrics are calculated from irrigation_data records matching the filters:

- **total_irrigation_volume_mm**: Sum of all `real_amount` values (actual water delivered)
- **total_irrigation_events**: Count of irrigation records
- **average_efficiency**: Average of (real_amount / nominal_amount) across all events
  - Excludes records where nominal_amount ≤ 0
  - Returns `null` if no valid efficiencies exist
- **efficiency_range**: Min and max efficiency across valid events
  - Returns `null` if no valid efficiencies exist

### Efficiency Calculation

```
efficiency = real_amount / nominal_amount
```

**Edge Cases:**
- If `nominal_amount` = 0 or negative, that record's efficiency is excluded from averaging (returns `null`)
- If `nominal_amount` > 0 but `real_amount` < 0, efficiency is calculated as negative (indicates data issue)
- If all records have invalid nominal_amounts, `average_efficiency` and `efficiency_range` are `null`

### Year-over-Year Comparison

For requested date range `[start_date, end_date]`:
- **same_period_-1**: Same calendar dates in previous year (e.g., Jan 1-31 of last year)
- **same_period_-2**: Same calendar dates two years ago (e.g., Jan 1-31 from 2 years ago)

**Missing Data Handling:**
- If no data exists for a previous year period, the comparison object contains:
  - All metrics as `null`
  - `data_incomplete: true`
  - `note: "No data available for {year description}"`
- Corresponding percentage changes in `period_comparison` are `null`

### Percentage Change Calculation

```
volume_change_percent = ((current_volume - previous_volume) / previous_volume) * 100
events_change_percent = ((current_events - previous_events) / previous_events) * 100
efficiency_change_percent = ((current_efficiency - previous_efficiency) / previous_efficiency) * 100
```

**Returns `null` if:**
- Previous period data is missing (`data_incomplete: true`)
- Previous period value is 0 or negative (division by zero prevention)
- Current or previous efficiency is `null`

### Time-Series Aggregation

Data is grouped by time bucket depending on aggregation type:

#### Daily (default)
- One entry per calendar day
- `date` format: `YYYY-MM-DD`
- `efficiency`: Average efficiency for that day
- Uses PostgreSQL `DATE_TRUNC('day', start_time)`

#### Weekly
- One entry per ISO week
- `date` format: Start date of week (YYYY-MM-DD)
- `efficiency`: Average efficiency for that week
- Uses PostgreSQL `DATE_TRUNC('week', start_time)`

#### Monthly
- One entry per calendar month
- `date` format: First day of month (YYYY-MM-DD)
- `efficiency`: Average efficiency for that month
- Uses PostgreSQL `DATE_TRUNC('month', start_time)`

**Database Optimization:**
All aggregations use SQL `GROUP BY DATE_TRUNC()` to push computation to PostgreSQL, leveraging composite indexes `(farm_id, start_time)` and `(irrigation_sector_id, start_time)` for optimal performance per DatabaseOptimization.md.

### Pagination

Time-series results are paginated to prevent large response payloads:

- **page**: 1-indexed page number
- **limit**: Results per page (1-1000, default 50)
- **total_count**: Total records matching filters (before pagination)
- **total_pages**: Calculated as `ceil(total_count / limit)`

To fetch all results, use `limit=all` (capped at 10,000 results). Caution: Very large datasets may exceed HTTP timeouts.

### Sector Breakdown

Aggregated metrics grouped by irrigation sector:

- If `sector_id` query param provided: Only that sector appears
- If `sector_id` omitted: All farm sectors included
- **total_volume_mm**: Sum of `real_amount` for the sector
- **average_efficiency**: Average efficiency for the sector (null if no valid data)

## Example Requests

### Daily analytics for a farm (last 90 days)
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics"
```

### Specific date range with daily aggregation
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics?start_date=2024-01-01&end_date=2024-01-31&aggregation=daily"
```

### Weekly aggregation for a specific sector
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics?sector_id=5&aggregation=weekly"
```

### Monthly analytics with pagination
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics?aggregation=monthly&page=2&limit=100"
```

### Get all results (careful with large datasets)
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics?limit=all"
```

## Timeout Considerations

### HTTP Timeouts
- **ReadTimeout**: 15 seconds (default)
- **WriteTimeout**: 15 seconds (default)
- **Analytics query timeout**: 30 seconds (specifically for analytics queries)

### Risk of Timeout
- **limit=all** with >100,000 records may exceed 15s WriteTimeout
- Large date ranges (>2 years) may slow queries
- Recommend using pagination (default limit=50) for production APIs

### Mitigation
- Use page-based pagination for browsing large datasets
- Limit date ranges to quarterly or monthly windows
- Monitor query performance with `EXPLAIN ANALYZE` on PostgreSQL

## Performance Notes

Per DatabaseOptimization.md best practices:

1. **Composite Index Usage**: Queries leverage `idx_irrigation_farm_time` and `idx_irrigation_sector_time` for efficient range scans
2. **SQL-Level Aggregation**: All GROUP BY and calculations occur at database level, not application level
3. **Single Query for YoY**: Year-over-year data fetched via single UNION ALL query for efficiency
4. **No N+1 Queries**: Relations (Farm, IrrigationSector) joined at query level, not fetched per-row

Example query execution time: <100ms for 1-year data range with aggregation on mid-scale farm (4 sectors, ~500 events).

## Troubleshooting

### Empty time_series
- Check that start_date/end_date don't fall outside irrigation season (typical: March 1 - October 31)
- Verify farm_id exists and has irrigation data in the date range
- Use `GET /health` to confirm API is operational

### null efficiency values
- Occurs when nominal_amount ≤ 0 in all records for that period
- Check data quality: nominal_amount should always be > 0
- This is a data issue in the source system, not an API issue

### 206 (Partial Content) responses
- Indicates previous year data is missing
- Normal for new farms with <2 years of history
- Year-over-year comparisons will be unavailable (`vs_same_period_-1` may be null)

### Timeouts on large "all" results
- Reduce limit or use pagination
- Narrow date range to most recent months
- Consider running analytics for specific sectors separately

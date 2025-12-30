# Test Cases for Irrigation Analytics API

This document outlines test scenarios validated during development and serves as a guide for implementing unit and integration tests.

## Integration Tests - Analytics Endpoint

### Test 1: Default Query (Last 90 Days)
**Endpoint:** `GET /v1/farms/:farm_id/irrigation/analytics`

**Test Command:**
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics"
```

**Expected Behavior:**
- HTTP 200 or 206 status code
- Default to last 90 days from current date
- Daily aggregation (default)
- Pagination: page=1, limit=50
- All response fields present (farm_id, period, metrics, yoy comparisons, time_series, sector_breakdown)

**Assertions:**
- Response contains all top-level keys: `farm_id`, `farm_name`, `period`, `aggregation`, `metrics`, `same_period_-1`, `same_period_-2`, `period_comparison`, `time_series`, `sector_breakdown`
- `aggregation` field equals "daily"
- `time_series.pagination.page` equals 1
- `time_series.pagination.limit` equals 50

---

### Test 2: Custom Date Range with Weekly Aggregation
**Endpoint:** `GET /v1/farms/:farm_id/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-31&aggregation=weekly`

**Test Command:**
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-31&aggregation=weekly"
```

**Expected Response Structure:**
```json
{
  "farm_id": 1,
  "period": {
    "start": "2024-03-01T00:00:00Z",
    "end": "2024-03-31T23:59:59.999999999Z"
  },
  "aggregation": "weekly",
  "metrics": {
    "total_irrigation_volume_mm": 766.74,
    "total_irrigation_events": 36,
    "average_efficiency": 0.889,
    "efficiency_range": {
      "min": 0.75,
      "max": 0.98
    }
  },
  "time_series": {
    "data": [/* weekly buckets */],
    "pagination": { /* metadata */ }
  }
}
```

**Assertions:**
- HTTP 200 status (data exists for 2024, 2023, 2022)
- `period.start` matches "2024-03-01T00:00:00Z"
- `period.end` matches "2024-03-31T23:59:59.999999999Z"
- `aggregation` equals "weekly"
- `time_series.data` contains 5 entries (4 full weeks + partial week at start)
- `time_series.data[0].date` equals "2024-02-25" (start of ISO week containing March 1)
- `metrics.total_irrigation_events` equals 36 (9 events/week × 4 weeks)
- YoY data present for both `-1` and `-2` periods
- `same_period_-1.data_incomplete` is false
- `same_period_-2.data_incomplete` is false

---

### Test 3: Monthly Aggregation
**Endpoint:** `GET /v1/farms/:farm_id/irrigation/analytics?start_date=2024-01-01&end_date=2024-03-31&aggregation=monthly`

**Test Command:**
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics?start_date=2024-01-01&end_date=2024-03-31&aggregation=monthly"
```

**Assertions:**
- `aggregation` equals "monthly"
- `time_series.data` contains 3 entries (Jan, Feb, Mar)
- `time_series.data[0].date` equals "2024-01-01"
- `time_series.data[1].date` equals "2024-02-01"
- `time_series.data[2].date` equals "2024-03-01"
- Each entry aggregates all irrigation events for that month

---

### Test 4: Sector Filtering
**Endpoint:** `GET /v1/farms/:farm_id/irrigation/analytics?sector_id=1`

**Test Command:**
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-31&sector_id=1"
```

**Expected Behavior:**
- Only sector 1 appears in `sector_breakdown` array
- `metrics` still includes all farm sectors (sector filter applies to breakdown only, not main metrics)

**Assertions:**
- `sector_breakdown` array has length 1
- `sector_breakdown[0].sector_id` equals 1
- `sector_breakdown[0].sector_name` equals "North Field"
- `sector_breakdown[0].total_volume_mm` > 0
- `sector_breakdown[0].average_efficiency` between 0.0 and 1.0

---

### Test 5: Pagination (Page 1)
**Endpoint:** `GET /v1/farms/:farm_id/irrigation/analytics?limit=5&page=1`

**Test Command:**
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-31&limit=5&page=1"
```

**Assertions:**
- `time_series.pagination.page` equals 1
- `time_series.pagination.limit` equals 5
- `time_series.data` array has length 5
- `time_series.pagination.total_count` > 5
- `time_series.pagination.total_pages` calculated as `ceil(total_count / limit)`

---

### Test 6: Pagination (Page 2)
**Endpoint:** `GET /v1/farms/:farm_id/irrigation/analytics?limit=5&page=2`

**Test Command:**
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-31&limit=5&page=2"
```

**Assertions:**
- `time_series.pagination.page` equals 2
- `time_series.data` array has length 5 (or less if near end)
- `time_series.data[0].date` is later than the last date from page 1

---

### Test 7: HTTP 200 - Complete YoY Data
**Endpoint:** `GET /v1/farms/:farm_id/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-31`

**Test Command:**
```bash
curl -o /dev/null -w "%{http_code}" "http://localhost:8080/v1/farms/1/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-31"
```

**Expected:**
- HTTP 200 status code (all YoY periods have data)

**Assertions:**
- Status code is 200
- `same_period_-1.data_incomplete` is false
- `same_period_-2.data_incomplete` is false
- `same_period_-1.total_irrigation_volume_mm` is not null
- `same_period_-2.total_irrigation_volume_mm` is not null
- `period_comparison.vs_same_period_-1` contains valid percentage values
- `period_comparison.vs_same_period_-2` contains valid percentage values

---

### Test 8: HTTP 206 - Incomplete YoY Data
**Endpoint:** `GET /v1/farms/:farm_id/irrigation/analytics?start_date=2023-01-01&end_date=2023-01-31`

**Test Command:**
```bash
curl -o /dev/null -w "%{http_code}" "http://localhost:8080/v1/farms/1/irrigation/analytics?start_date=2023-01-01&end_date=2023-01-31"
```

**Expected:**
- HTTP 206 status code (partial content - previous years missing)

**Test for Data Incompleteness:**
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics?start_date=2023-01-01&end_date=2023-01-31" | jq '.["same_period_-1"]'
```

**Assertions:**
- Status code is 206
- `same_period_-1.data_incomplete` is true
- `same_period_-1.note` contains "No data available for previous year"
- `same_period_-1.total_irrigation_volume_mm` is null
- `same_period_-1.total_irrigation_events` is null
- `same_period_-1.average_efficiency` is null
- `period_comparison.vs_same_period_-1.volume_change_percent` is null

---

### Test 9: HTTP 400 - Invalid Date Format
**Endpoint:** `GET /v1/farms/:farm_id/irrigation/analytics?start_date=invalid`

**Test Command:**
```bash
curl -o /dev/null -w "%{http_code}" "http://localhost:8080/v1/farms/1/irrigation/analytics?start_date=invalid"
```

**Expected:**
- HTTP 400 status code
- Error message indicating invalid date format

**Assertions:**
- Status code is 400
- Response body contains `"error"` key
- Error message mentions "date format" or "YYYY-MM-DD"

---

### Test 10: HTTP 400 - Invalid Aggregation Type
**Endpoint:** `GET /v1/farms/:farm_id/irrigation/analytics?aggregation=yearly`

**Test Command:**
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics?aggregation=yearly"
```

**Expected:**
- HTTP 400 status code
- Error message indicating invalid aggregation type

**Assertions:**
- Status code is 400
- Error message mentions "aggregation" and valid values (daily/weekly/monthly)

---

### Test 11: HTTP 404 - Farm Not Found
**Endpoint:** `GET /v1/farms/99999/irrigation/analytics`

**Test Command:**
```bash
curl -o /dev/null -w "%{http_code}" "http://localhost:8080/v1/farms/99999/irrigation/analytics"
```

**Expected:**
- HTTP 404 status code
- Error message indicating farm not found

---

### Test 12: Efficiency Calculation Validation
**Endpoint:** `GET /v1/farms/:farm_id/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-05`

**Manual Verification:**
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-05" | jq '.time_series.data[0]'
```

**Assertions:**
- For each time series entry: `efficiency = real_amount_mm / nominal_amount_mm`
- Efficiency values are between 0.0 and 1.0 (or slightly above for data quality issues)
- If nominal_amount_mm is 0, efficiency should be null (excluded from calculation)

---

### Test 13: Empty Result Set (No Data)
**Endpoint:** `GET /v1/farms/:farm_id/irrigation/analytics?start_date=2020-01-01&end_date=2020-01-31`

**Test Command:**
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics?start_date=2020-01-01&end_date=2020-01-31"
```

**Expected Behavior:**
- HTTP 200 status code (request is valid)
- Empty or zero-value metrics

**Assertions:**
- `metrics.total_irrigation_volume_mm` equals 0 or null
- `metrics.total_irrigation_events` equals 0
- `time_series.data` is empty array
- `sector_breakdown` is empty array

---

### Test 14: Percentage Change Calculation
**Endpoint:** `GET /v1/farms/:farm_id/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-31`

**Test Command:**
```bash
curl "http://localhost:8080/v1/farms/1/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-31" | jq '.period_comparison'
```

**Manual Calculation:**
```
volume_change_percent = ((current_volume - previous_volume) / previous_volume) * 100
```

**Assertions:**
- Percentage values are correctly calculated
- Positive values indicate increase
- Negative values indicate decrease
- If previous period has 0 volume, percentage should be null

---

## Unit Tests - Repository Layer

### IrrigationDataRepository.GetAnalyticsForFarmByDateRange

**Test Cases:**
1. **Valid query with daily aggregation**
   - Input: farmID=1, startDate="2024-03-01", endDate="2024-03-31", aggregation="daily", limit=50, offset=0
   - Expected: Array of AnalyticsAggregation structs with daily buckets
   - Verify: DATE_TRUNC('day', start_time) used in SQL

2. **Weekly aggregation**
   - Input: aggregation="weekly"
   - Expected: Buckets start on Mondays (ISO week standard)
   - Verify: DATE_TRUNC('week', start_time) used

3. **Monthly aggregation**
   - Input: aggregation="monthly"
   - Expected: Buckets start on 1st of each month
   - Verify: DATE_TRUNC('month', start_time) used

4. **Efficiency calculation with zero nominal_amount**
   - Setup: Insert record with nominal_amount=0
   - Expected: Efficiency is null for that record
   - Verify: CASE WHEN nominal_amount > 0 handles edge case

5. **Pagination offset calculation**
   - Input: page=2, limit=10 → offset=10
   - Expected: Correct offset applied in SQL LIMIT/OFFSET
   - Verify: Results skip first 10 records

6. **Total count accuracy**
   - Expected: totalCount matches actual number of aggregated buckets
   - Verify: Count query returns correct value

### IrrigationDataRepository.GetYoYComparison

**Test Cases:**
1. **All three years have data**
   - Input: farmID=1, startDate="2024-03-01", endDate="2024-03-31"
   - Expected: Map with keys for years 2024, 2023, 2022
   - Verify: All three map entries have non-zero values

2. **Missing previous year data**
   - Input: startDate="2023-03-01", endDate="2023-03-31"
   - Expected: Map entries for 2023, 2022, 2021 (with 2022/2021 possibly empty)
   - Verify: Empty years return zero values, not errors

3. **SQL UNION ALL efficiency**
   - Verify: Single query with UNION ALL, not 3 separate queries
   - Verify: Query performance <100ms for typical dataset

4. **Date offset calculation**
   - Input: 2024-03-01 to 2024-03-31
   - Expected Year -1: 2023-03-01 to 2023-03-31
   - Expected Year -2: 2022-03-01 to 2022-03-31
   - Verify: Date arithmetic correct (365 days × offset)

### IrrigationDataRepository.GetSectorBreakdownForFarm

**Test Cases:**
1. **All sectors in farm**
   - Input: farmID=1, sectorID=nil
   - Expected: Array with all 4 sectors (North Field, South Field, East Pasture, West Pasture)
   - Verify: JOIN to irrigation_sectors table retrieves sector names

2. **Single sector filter**
   - Input: farmID=1, sectorID=1
   - Expected: Array with 1 entry (sector_id=1)
   - Verify: WHERE clause filters correctly

3. **Efficiency aggregation**
   - Expected: average_efficiency = AVG(real_amount / nominal_amount)
   - Verify: Null handling for zero nominal_amount

4. **Empty result (no data for sector)**
   - Input: sectorID that has no irrigation events
   - Expected: Empty array or zero values
   - Verify: No database errors

---

## Unit Tests - Service Layer

### IrrigationAnalyticsService.GetAnalytics

**Test Cases:**
1. **Default date range (nil start/end dates)**
   - Input: startDate=nil, endDate=nil
   - Expected: Defaults to last 90 days
   - Verify: Service calculates `time.Now().AddDate(0, 0, -90)`

2. **Orchestration of repository calls**
   - Mock all repository methods
   - Verify: GetAnalyticsForFarmByDateRange called once
   - Verify: GetYoYComparison called once
   - Verify: GetSectorBreakdownForFarm called once

3. **Error handling from repository**
   - Mock: GetAnalyticsForFarmByDateRange returns error
   - Expected: Service returns error (doesn't panic)
   - Verify: Error is wrapped with context

4. **Null efficiency handling**
   - Input: Data with all nominal_amount=0
   - Expected: average_efficiency is nil
   - Verify: No division by zero errors

### IrrigationAnalyticsService.calculateMetrics

**Test Cases:**
1. **Aggregate multiple records**
   - Input: []AnalyticsAggregation with 3 daily entries
   - Expected: Sum of volumes, sum of events, average efficiency
   - Verify: Calculations are accurate

2. **Empty input array**
   - Input: []AnalyticsAggregation{}
   - Expected: Zero values, nil efficiency
   - Verify: No panic

3. **Efficiency range calculation**
   - Input: Data with efficiencies [0.75, 0.88, 0.95, 0.98]
   - Expected: min=0.75, max=0.98
   - Verify: Correct min/max values

### IrrigationAnalyticsService.calculatePercentageChanges

**Test Cases:**
1. **Positive change**
   - Input: current=100, previous=80
   - Expected: 25.0 (25% increase)

2. **Negative change**
   - Input: current=80, previous=100
   - Expected: -20.0 (20% decrease)

3. **Zero previous value**
   - Input: current=100, previous=0
   - Expected: nil (avoid division by zero)

4. **Null efficiency values**
   - Input: currentEfficiency=nil or previousEfficiency=nil
   - Expected: efficiency_change_percent=nil

---

## Unit Tests - Controller Layer

### AnalyticsController.GetAnalytics

**Test Cases:**
1. **Parse farm_id from path**
   - Input: URL "/v1/farms/123/irrigation/analytics"
   - Expected: farmID=123 passed to service

2. **Parse query parameters**
   - Input: "?start_date=2024-03-01&end_date=2024-03-31&aggregation=weekly&page=2&limit=100"
   - Expected: All params correctly parsed and passed to service

3. **Date parsing error**
   - Input: "?start_date=invalid"
   - Expected: HTTP 400, error message about date format

4. **Invalid aggregation type**
   - Input: "?aggregation=yearly"
   - Expected: HTTP 400, error message listing valid types

5. **Invalid page number**
   - Input: "?page=0" or "?page=-1"
   - Expected: HTTP 400, error message about valid page numbers

6. **Limit "all" conversion**
   - Input: "?limit=all"
   - Expected: limit=10000 passed to service

7. **HTTP status code determination**
   - Mock service response with data_incomplete=false for both YoY periods
   - Expected: HTTP 200
   - Mock service response with data_incomplete=true for one YoY period
   - Expected: HTTP 206

8. **Service error handling**
   - Mock service returns error
   - Expected: HTTP 500, error message in response

---

## Integration Test Scenarios (End-to-End)

### Scenario 1: Full Analytics Workflow
1. Seed database with known data (2023-2025)
2. Request analytics for March 2024
3. Verify all calculations match expected values
4. Verify YoY comparisons are accurate
5. Verify sector breakdowns sum to total

### Scenario 2: Pagination Consistency
1. Request page 1 with limit=10
2. Request page 2 with limit=10
3. Request limit=all
4. Verify: Records from page1 + page2 match first 20 records from limit=all

### Scenario 3: Aggregation Consistency
1. Request daily aggregation for March 2024
2. Request weekly aggregation for March 2024
3. Request monthly aggregation for March 2024
4. Verify: Sum of daily volumes = sum of weekly volumes = monthly volume

### Scenario 4: Concurrency Test
1. Send 10 simultaneous requests to analytics endpoint
2. Verify: All return consistent results
3. Verify: No database connection pool exhaustion

### Scenario 5: Performance Test
1. Load 100,000+ irrigation records
2. Request analytics with limit=all
3. Verify: Response time <2 seconds
4. Verify: No timeout errors

---

## Edge Cases to Test

1. **Leap year handling**
   - Date range includes Feb 29 on leap year vs. non-leap year
   - Verify YoY comparison handles date offset correctly

2. **Timezone boundaries**
   - All dates interpreted as UTC
   - Verify start_date=2024-03-01 starts at 00:00:00 UTC

3. **Very large efficiency values**
   - Scenario: real_amount > nominal_amount (system inefficiency/measurement error)
   - Expected: Efficiency > 1.0 (don't clamp, indicates data quality issue)

4. **Negative amounts**
   - Scenario: real_amount < 0 (data error)
   - Expected: Negative efficiency or excluded from calculation

5. **Extremely large date ranges**
   - Input: 10-year date range
   - Expected: Pagination prevents timeout, total_count accurate

6. **Farm with no sectors**
   - Expected: Empty sector_breakdown array
   - Verify: No errors, graceful handling

7. **Sector filter with invalid sector_id**
   - Input: sector_id=99999 (doesn't exist)
   - Expected: Empty sector_breakdown array
   - Verify: HTTP 200 (valid request, just no matching data)

---

## Performance Benchmarks

Track these metrics during integration testing:

1. **Query Performance**
   - GetAnalyticsForFarmByDateRange: Target <50ms for 1-year range
   - GetYoYComparison: Target <100ms (UNION ALL of 3 queries)
   - GetSectorBreakdownForFarm: Target <30ms

2. **Memory Usage**
   - Limit=1000: Peak memory <50MB per request
   - Limit=all (10,000 records): Peak memory <200MB

3. **Database Connection Pool**
   - Verify connections released after request
   - No connection leaks under concurrent load

4. **Response Size**
   - Limit=50: Response size <100KB
   - Limit=1000: Response size <1MB

---

## Test Data Requirements

For comprehensive testing, seed database with:

1. **Multiple Farms:** At least 2 farms
2. **Multiple Sectors:** 4+ sectors per farm
3. **Time Coverage:** 3+ years of data (2023-2025 minimum)
4. **Seasonal Variation:** Data across all months (not just growing season)
5. **Efficiency Profiles:**
   - Excellent (>95%)
   - Good (85-95%)
   - Average (75-85%)
   - Poor (<75%)
6. **Edge Cases:**
   - Records with nominal_amount=0
   - Records with real_amount > nominal_amount
   - Gaps in data (missing weeks/months)
   - Single-day events vs. multi-day periods

---

## CI/CD Integration

Recommended test execution order:

1. **Unit Tests** (repository, service, controller layers)
   - Run on every commit
   - Must pass before merge

2. **Integration Tests** (API endpoints with test database)
   - Run on pull requests
   - Seed test database with known data
   - Verify all 14+ endpoint scenarios

3. **Performance Tests** (load testing)
   - Run nightly or pre-release
   - Validate response times under load
   - Check for memory leaks

4. **Regression Tests** (validate existing functionality)
   - Run before production deployment
   - Ensure backward compatibility

---

## Notes

- All test commands assume server running at `http://localhost:8080`
- Tests validated against seed data in `internal/seeds/irrigation_seed.json`
- Date ranges use YYYY-MM-DD format (ISO 8601)
- HTTP status codes follow REST best practices (200, 206, 400, 404, 500)
- Efficiency calculation: `real_amount / nominal_amount` with null handling

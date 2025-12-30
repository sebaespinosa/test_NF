#!/usr/bin/env bash
set -uo pipefail

BASE_URL=${BASE_URL:-http://localhost:8080}
FARM_ID=${FARM_ID:-1}
SECTOR_ID=${SECTOR_ID:-1}

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required for integration checks. Install jq and retry." >&2
  exit 1
fi

successes=()
failures=()

die() {
  echo "$1" >&2
  exit 1
}

run_case() {
  local name="$1"
  local url="$2"
  local expected_statuses="$3"
  local jq_assertion="$4"

  local tmp status
  tmp=$(mktemp)
  status=$(curl -s -o "$tmp" -w "%{http_code}" "$url") || status="curl_error"

  local ok=false
  IFS=',' read -ra expected <<< "$expected_statuses"
  for e in "${expected[@]}"; do
    if [[ "$status" == "$e" ]]; then
      ok=true
      break
    fi
  done

  if [[ "$ok" != true ]]; then
    failures+=("$name (status=$status expected=$expected_statuses)")
    rm -f "$tmp"
    echo "[FAIL] $name (status=$status)"
    return
  fi

  if [[ -n "$jq_assertion" ]]; then
    if ! jq -e "$jq_assertion" "$tmp" >/dev/null 2>&1; then
      failures+=("$name (jq assertion failed)")
      rm -f "$tmp"
      echo "[FAIL] $name (jq assertion failed)"
      return
    fi
  fi

  successes+=("$name")
  rm -f "$tmp"
  echo "[PASS] $name"
}

echo "Running integration checks against $BASE_URL (farm=$FARM_ID, sector=$SECTOR_ID)"

run_case "Default 90d" \
  "$BASE_URL/v1/farms/$FARM_ID/irrigation/analytics" \
  "200,206" \
  '.aggregation=="daily" and .time_series.pagination.page==1 and .time_series.pagination.limit==50'

run_case "Weekly range" \
  "$BASE_URL/v1/farms/$FARM_ID/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-31&aggregation=weekly" \
  "200" \
  '.aggregation=="weekly" and (.time_series.data|length)==5'

run_case "Monthly range" \
  "$BASE_URL/v1/farms/$FARM_ID/irrigation/analytics?start_date=2024-01-01&end_date=2024-03-31&aggregation=monthly" \
  "200" \
  '.aggregation=="monthly" and (.time_series.data|length)>=1'

run_case "Sector filter" \
  "$BASE_URL/v1/farms/$FARM_ID/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-31&sector_id=$SECTOR_ID" \
  "200" \
  '(.sector_breakdown|length)==1 and (.sector_breakdown[0].sector_id=='"$SECTOR_ID"')'

run_case "Pagination p1" \
  "$BASE_URL/v1/farms/$FARM_ID/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-31&limit=5&page=1" \
  "200" \
  '.time_series.pagination.page==1 and .time_series.pagination.limit==5 and (.time_series.data|length)==5'

run_case "Pagination p2" \
  "$BASE_URL/v1/farms/$FARM_ID/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-31&limit=5&page=2" \
  "200" \
  '.time_series.pagination.page==2 and .time_series.pagination.limit==5 and (.time_series.data|length)>=1'

run_case "YoY complete (200)" \
  "$BASE_URL/v1/farms/$FARM_ID/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-31" \
  "200" \
  '."same_period_-1".data_incomplete==false and ."same_period_-2".data_incomplete==false'

run_case "YoY incomplete (206)" \
  "$BASE_URL/v1/farms/$FARM_ID/irrigation/analytics?start_date=2023-01-01&end_date=2023-01-31" \
  "206" \
  '."same_period_-1".data_incomplete==true'

run_case "Invalid date" \
  "$BASE_URL/v1/farms/$FARM_ID/irrigation/analytics?start_date=invalid" \
  "400" \
  ''

run_case "Invalid aggregation" \
  "$BASE_URL/v1/farms/$FARM_ID/irrigation/analytics?aggregation=yearly" \
  "400" \
  ''

run_case "Empty range" \
  "$BASE_URL/v1/farms/$FARM_ID/irrigation/analytics?start_date=2020-01-01&end_date=2020-01-31" \
  "200,206" \
  '(.time_series.data|length)==0'

run_case "Percentage change" \
  "$BASE_URL/v1/farms/$FARM_ID/irrigation/analytics?start_date=2024-03-01&end_date=2024-03-31" \
  "200,206" \
  '(.period_comparison."vs_same_period_-1".volume_change_percent != null) or (.period_comparison."vs_same_period_-2".volume_change_percent != null)'

echo
echo "Summary: ${#successes[@]} passed, ${#failures[@]} failed"
if [[ ${#failures[@]} -gt 0 ]]; then
  printf 'Failures:\n'
  for f in "${failures[@]}"; do
    printf ' - %s\n' "$f"
  done
  exit 1
fi
exit 0

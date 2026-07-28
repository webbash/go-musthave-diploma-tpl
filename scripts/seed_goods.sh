#!/usr/bin/env bash
set -euo pipefail

ACCRUAL_URL="${1:-${ACCRUAL_URL:-http://localhost:8081}}"

create_good() {
  local match="$1"
  local reward="$2"
  local reward_type="$3"
  local status

  status="$(
    curl -sS -o /dev/null -w '%{http_code}' -X POST "${ACCRUAL_URL}/api/goods" \
    -H "Content-Type: application/json" \
    -d "{
      \"match\": \"${match}\",
      \"reward\": ${reward},
      \"reward_type\": \"${reward_type}\"
    }"
  )"

  case "${status}" in
    200|201)
      echo "Created good: ${match}"
      ;;
    409)
      echo "Good already exists: ${match}"
      ;;
    *)
      echo "Failed to create good ${match}: HTTP ${status}" >&2
      return 1
      ;;
  esac
}

create_good "Bork" 10 "%"
create_good "Philips" 5 "%"
create_good "чай" 2 "%"
create_good "Philips HD" 7 "%"

echo "Finished seeding goods in ${ACCRUAL_URL}"

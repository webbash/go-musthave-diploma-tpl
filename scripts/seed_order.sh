#!/usr/bin/env bash
set -euo pipefail

ACCRUAL_URL="${1:-${ACCRUAL_URL:-http://localhost:8081}}"
ORDER_NUMBER="${2:-${ORDER_NUMBER:-}}"

generate_order_number() {
  local base=""
  local i digit sum=0 double=1 check

  for ((i = 0; i < 15; i++)); do
    digit=$((RANDOM % 10))
    base+="${digit}"
  done

  for ((i = 14; i >= 0; i--)); do
    digit=${base:i:1}
    if (( double )); then
      digit=$((digit * 2))
      if (( digit > 9 )); then
        digit=$((digit - 9))
      fi
    fi
    sum=$((sum + digit))
    double=$((1 - double))
  done

  check=$(((10 - sum % 10) % 10))
  echo "${base}${check}"
}

if [[ -z "${ORDER_NUMBER}" ]]; then
  ORDER_NUMBER="$(generate_order_number)"
fi

echo "Using order number: ${ORDER_NUMBER}"

status="$(
  curl -sS -o /dev/null -w '%{http_code}' -X POST "${ACCRUAL_URL}/api/orders" \
  -H "Content-Type: application/json" \
  -d "{
    \"order\": \"${ORDER_NUMBER}\",
    \"goods\": [
      {\"description\": \"Чайник Bork\", \"price\": 7000},
      {\"description\": \"Тостер Bork\", \"price\": 4500},
      {\"description\": \"Утюг Philips\", \"price\": 3200},
      {\"description\": \"Пакет чая\", \"price\": 120},
      {\"description\": \"Фильтр для воды\", \"price\": 890}
    ]
  }"
)"

case "${status}" in
  200|201|202)
    echo "Created order: ${ORDER_NUMBER}"
    ;;
  409)
    echo "Order already exists: ${ORDER_NUMBER}"
    ;;
  *)
    echo "Failed to create order ${ORDER_NUMBER}: HTTP ${status}" >&2
    exit 1
    ;;
esac

echo "Created test order ${ORDER_NUMBER} in ${ACCRUAL_URL}"

#!/usr/bin/env bash
set -euo pipefail

COUNT="${1:-500}"
API_BASE="${API_BASE:-http://localhost:8080}"
API_V1="${API_BASE%/}/api/v1"

USERS=(
  "alice"
  "bob"
  "charlie"
  "dana"
  "eli"
  "frankie"
  "gina"
  "hugo"
  "ivy"
  "janky"
  "matt"
  "nora"
)

DESCRIPTIONS=(
  "dinner"
  "groceries"
  "rent split"
  "utilities"
  "coffee run"
  "uber"
  "movie tickets"
  "concert"
  "lunch"
  "weekend trip"
)

if ! [[ "$COUNT" =~ ^[0-9]+$ ]] || [[ "$COUNT" -le 0 ]]; then
  echo "COUNT must be a positive integer. Got: $COUNT" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required. Install jq and re-run this script." >&2
  exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required. Install curl and re-run this script." >&2
  exit 1
fi

random_amount() {
  # 1.00 to 500.00
  awk -v seed="$RANDOM" 'BEGIN { srand(seed); printf "%.2f", (1 + rand() * 499) }'
}

pick_currency_pair() {
  local r=$((RANDOM % 100))
  if (( r < 40 )); then
    echo "USD USD"
  elif (( r < 60 )); then
    echo "JPY EUR"
  elif (( r < 85 )); then
    echo "JPY USD"
  elif (( r < 95 )); then
    echo "EUR USD"
  else
    echo "USD JPY"
  fi
}

pick_owers() {
  local payer="$1"
  local max_owers=3
  local ower_count=$((1 + RANDOM % max_owers))
  local picked=()
  declare -A seen=()

  while (( ${#picked[@]} < ower_count )); do
    local candidate="${USERS[$((RANDOM % ${#USERS[@]}))]}"
    if [[ "$candidate" == "$payer" ]]; then
      continue
    fi
    if [[ -n "${seen[$candidate]:-}" ]]; then
      continue
    fi
    seen["$candidate"]=1
    picked+=("$candidate")
  done

  printf '%s\n' "${picked[@]}" | jq -R . | jq -s .
}

echo "API: $API_V1"
echo "Target payment count: $COUNT"
echo "Seeding users..."

for u in "${USERS[@]}"; do
  payload="$(jq -n --arg name "$u" '{name: $name}')"
  # Ignore duplicates/errors to keep the script idempotent-ish.
  curl -sS -o /dev/null -X POST \
    -H "Content-Type: application/json" \
    -d "$payload" \
    "$API_V1/user" || true
done

success=0
failed=0

echo "Injecting payments..."
for ((i = 1; i <= COUNT; i++)); do
  payer_idx=$((RANDOM % ${#USERS[@]}))
  payer="${USERS[$payer_idx]}"

  amount="$(random_amount)"
  desc="${DESCRIPTIONS[$((RANDOM % ${#DESCRIPTIONS[@]}))]}"
  read -r from_currency to_currency <<<"$(pick_currency_pair)"
  owers_json="$(pick_owers "$payer")"

  payload="$(
    jq -n \
      --arg payer "$payer" \
      --arg description "$desc #$i" \
      --arg from "$from_currency" \
      --arg to "$to_currency" \
      --argjson amount "$amount" \
      --argjson owers "$owers_json" \
      '{
        amount: $amount,
        payer: $payer,
        description: $description,
        fromExchangeRate: $from,
        toExchangeRate: $to,
        owers: $owers
      }'
  )"

  code=$(curl -sS -o /tmp/bsw_seed_resp.json -w "%{http_code}" \
    -X POST \
    -H "Content-Type: application/json" \
    -d "$payload" \
    "$API_V1/payment")

  if [[ "$code" == "200" ]]; then
    success=$((success + 1))
  else
    failed=$((failed + 1))
    msg="$(cat /tmp/bsw_seed_resp.json 2>/dev/null || true)"
    echo "[$i/$COUNT] failed (HTTP $code): $msg"
  fi

  if (( i % 50 == 0 )); then
    echo "Progress: $i/$COUNT (ok=$success, failed=$failed)"
  fi
done

echo "Done. ok=$success failed=$failed total=$COUNT"

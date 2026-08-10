#!/usr/bin/env bash
# Account 1001 dynamic leverage test — snapshot helper
set -euo pipefail

BASE="${BASE:-http://127.0.0.1:8080}"
MONLOG="${HOME}/.hstdev/logs/leverage-test-1001.log"
CORELOG="${HOME}/.hstdev/logs/hst-core.log"

snapshot() {
  local label="$1"
  local ts
  ts=$(date '+%Y-%m-%d %H:%M:%S')
  local svc_s svc_c svc_q
  svc_s=$(pgrep -x hst-server >/dev/null && echo UP || echo DOWN)
  svc_c=$(pgrep -x hst-core >/dev/null && echo UP || echo DOWN)
  svc_q=$(pgrep -x hst-quote >/dev/null && echo UP || echo DOWN)

  local acct pos tick bid
  acct=$(docker exec hst-server-postgres-1 psql -U hst -d hst -t -A -F'|' -c \
    "SELECT balance, equity, margin, margin_free, margin_level FROM accounts WHERE login=1001;" 2>/dev/null || echo "||||")
  pos=$(docker exec hst-server-postgres-1 psql -U hst -d hst -t -A -c \
    "SELECT count(*), coalesce(sum(volume),0) FROM positions WHERE login=1001;" 2>/dev/null || echo "0|0")
  bid=$(docker exec hst-server-redis-1 redis-cli GET 'hstquote:last:EURUSD' 2>/dev/null \
    | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('bid','-'))" 2>/dev/null || echo "-")

  local logline
  logline=$(grep -E 'resettled|refreshed|margin call|stop.?out|Position closed|login=1001|login\": 1001' "$CORELOG" 2>/dev/null | tail -3 | tr '\n' ' ')

  local line="[$ts] [$label] svc=s:$svc_s,c:$svc_c,q:$svc_q acct=$acct pos=$pos eurusd_bid=$bid logs=${logline:-none}"
  echo "$line" | tee -a "$MONLOG"
}

get_trader_token() {
  curl -sf -u '1001:Demo-Test-2026!' -X POST "$BASE/auth/trader/v1/login" \
    -H 'Content-Type: application/json' -d '{"connection_type":1}' \
    | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['access_token'])"
}

get_admin_token() {
  curl -sf -u '1000:Bootstrap-Admin-2026!' -X POST "$BASE/auth/v1/login" \
    -H 'Content-Type: application/json' -d '{"connection_type":32}' \
    | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['access_token'])"
}

case "${1:-snapshot}" in
  snapshot) snapshot "${2:-tick}" ;;
  token-trader) get_trader_token ;;
  token-admin) get_admin_token ;;
  *) echo "usage: $0 snapshot|token-trader|token-admin [label]" ;;
esac

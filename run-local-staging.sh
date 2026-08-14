#!/bin/bash
# Run and monitor the hst-v2 platform on a development machine.
#
# Each service is built into .service-bin and run with its own .env, exactly as
# the Makefile does, so there is one story for how a service is configured. What
# this adds on top is the monitoring surface: readiness polling against the real
# probe endpoints, per-service lifecycle, and a tick-feed liveness check.
#
# Infrastructure (postgres, redis, nats, influx) is not started here. Bring it up
# with brew services, or with `cd hst-server && make up` for the docker stack.
set -eo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"

BIN_DIR="$ROOT/.service-bin"
LOG_DIR="$ROOT/.service-logs"
mkdir -p "$BIN_DIR" "$LOG_DIR"

# ── Config ────────────────────────────────────────────────────────────────────
# Host ports for the infrastructure every service talks to. Override any of them
# from the environment; the services read their own .env, so these are only used
# for the health checks this script performs.
PG_HOST="${PG_HOST:-localhost}"
PG_PORT="${PG_PORT:-5432}"
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"
NATS_HOST="${NATS_HOST:-localhost}"
NATS_PORT="${NATS_PORT:-4222}"
NATS_MON_PORT="${NATS_MON_PORT:-8222}"
INFLUX_HOST="${INFLUX_HOST:-localhost}"
INFLUX_PORT="${INFLUX_PORT:-8086}"

# How long to wait for one service to answer its probe before giving up.
READY_TIMEOUT="${READY_TIMEOUT:-40}"

# Dependency services to health-check:  label:host:port:required|optional
DEPS=(
  "postgres:${PG_HOST}:${PG_PORT}:required"
  "redis:${REDIS_HOST}:${REDIS_PORT}:required"
  "nats:${NATS_HOST}:${NATS_PORT}:required"
  "influx:${INFLUX_HOST}:${INFLUX_PORT}:optional"
)

# The platform's own services, in start order. hst-server goes first: it runs the
# migrations, seeds, and serves the config that hst-quote and hst-news pull.
SERVICES=(
  "hst-server"
  "hst-core"
  "hst-quote"
  "hst-news"
)

# ── Reading a service's own .env ──────────────────────────────────────────────
# The .env file is the source of truth for a port, so the script reads it rather
# than keeping a second copy that can drift.
env_val() {
  local file="$ROOT/$1/.env" key="$2" def="$3" line
  [ -f "$file" ] || { echo "$def"; return 0; }

  line="$(grep -E "^[[:space:]]*${key}=" "$file" 2>/dev/null | tail -1)" || true
  [ -n "$line" ] || { echo "$def"; return 0; }

  line="${line#*=}"
  # drop a trailing comment, surrounding quotes and trailing spaces
  line="$(printf '%s' "$line" \
    | sed -e 's/[[:space:]]*#.*$//' -e 's/^"//' -e 's/"$//' -e "s/^'//" -e "s/'\$//" \
          -e 's/[[:space:]]*$//')"

  [ -n "$line" ] && echo "$line" || echo "$def"
}

# svc_port is the port a service answers its probe on.
svc_port() {
  case "$1" in
    hst-server) env_val hst-server HTTP_PORT 8080 ;;
    hst-core)   env_val hst-core   HEALTH_PORT 8082 ;;
    hst-quote)  env_val hst-quote  HEALTH_PORT 8081 ;;
    # .env ships 8082, which hst-core already holds; see health_port_overrides
    hst-news)   echo "${HST_NEWS_HEALTH_PORT:-8083}" ;;
    *) return 1 ;;
  esac
}

# svc_probe is the path that answers 200 once the service is up.
svc_probe() {
  case "$1" in
    hst-server) echo "/api/v1/system/monitor/health" ;;
    *)          echo "/healthz" ;;
  esac
}

# health_port_overrides are the env vars a service is started with, on top of its
# own .env. hst-core and hst-news both ship HEALTH_PORT=8082, and a bound port
# makes either one refuse to boot, so hst-news is moved out of the way.
health_port_overrides() {
  case "$1" in
    hst-news) echo "HEALTH_PORT=$(svc_port hst-news)" ;;
    *) : ;;
  esac
}

# Warn once if the .env files still carry the collision this script works around.
warn_health_port_clash() {
  local core news
  core="$(env_val hst-core HEALTH_PORT 8082)"
  news="$(env_val hst-news HEALTH_PORT 8082)"
  [ "$core" = "$news" ] || return 0
  printf "  \033[33m!\033[0m hst-core and hst-news both set HEALTH_PORT=%s in .env\n" "$core"
  printf "    starting hst-news on %s instead; fix hst-news/.env to make it stick\n" "$(svc_port hst-news)"
}

# ── Aliases ───────────────────────────────────────────────────────────────────
full_name() {
  case "$1" in
    server|api|hst-server) echo hst-server ;;
    core|engine|hst-core)  echo hst-core ;;
    quote|quotes|hst-quote) echo hst-quote ;;
    news|hst-news)         echo hst-news ;;
    *) return 1 ;;
  esac
}

# ── Generic helpers ───────────────────────────────────────────────────────────
# Build one service. ./cmd is the entry point in every module.
build_one() {
  local name="$1"
  ( cd "$ROOT/$name" && go build -o "$BIN_DIR/$name" ./cmd )
  echo "    built $name"
}

build_all() {
  echo "==> Building services..."
  local name
  for name in "${SERVICES[@]}"; do
    build_one "$name"
  done
  echo ""
}

# Start one service from its own directory, with its own .env.
#
# exec makes the subshell become the service, so the pidfile holds the real
# process and stopping it needs no walk down a process tree. nohup keeps it alive
# when the terminal that started it closes.
start_service() {
  local name="$1"; shift
  local dir="$ROOT/$name"

  if [ ! -x "$BIN_DIR/$name" ]; then
    printf "  \033[31m✗\033[0m %s: not built — run '%s build'\n" "$name" "$0" >&2
    return 1
  fi
  if [ ! -f "$dir/.env" ]; then
    printf "  \033[31m✗\033[0m %s: no .env — copy %s/.env.example first\n" "$name" "$name" >&2
    return 1
  fi

  if is_running "$name"; then
    echo "    $name already running (pid $(cat "$LOG_DIR/$name.pid"))"
    return 0
  fi

  echo "==> Starting $name..."
  (
    cd "$dir"
    set -a; . ./.env; set +a
    # overrides win over the service's own .env
    for kv in "$@"; do export "$kv"; done
    exec nohup "$BIN_DIR/$name"
  ) > "$LOG_DIR/$name.log" 2>&1 &

  echo $! > "$LOG_DIR/$name.pid"
  echo "    PID $! | logs: $LOG_DIR/$name.log"
}

# is_running reports whether the pidfile names a live process.
is_running() {
  local pidfile="$LOG_DIR/$1.pid"
  [ -f "$pidfile" ] || return 1
  kill -0 "$(cat "$pidfile")" 2>/dev/null
}

# Kill whatever is listening on a TCP port, for an orphan the pidfile missed.
free_port() {
  local port="$1" pids
  [ -n "$port" ] || return 0
  pids="$(lsof -nP -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null)" || true
  [ -n "$pids" ] || return 0
  kill $pids 2>/dev/null || true
  sleep 1
  pids="$(lsof -nP -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null)" || true
  [ -n "$pids" ] && kill -9 $pids 2>/dev/null || true
}

# probe_ok asks the service whether it is up.
probe_ok() {
  local name="$1" port path
  port="$(svc_port "$name")" || return 1
  path="$(svc_probe "$name")"
  curl -sf -m 2 -o /dev/null "http://localhost:${port}${path}" 2>/dev/null
}

# Wait for one service to answer its probe. Notices a crash rather than waiting it out.
wait_service() {
  local name="$1" tries="${2:-$READY_TIMEOUT}" n=0
  while [ "$n" -lt "$tries" ]; do
    if probe_ok "$name"; then
      printf "  \033[32m✓\033[0m %-11s ready on :%s\n" "$name" "$(svc_port "$name")"
      return 0
    fi
    # the process exiting is a failed boot, not a slow one
    if ! is_running "$name"; then
      printf "  \033[31m✗\033[0m %-11s exited during startup — see %s\n" \
        "$name" "$LOG_DIR/$name.log"
      tail -n 3 "$LOG_DIR/$name.log" 2>/dev/null | sed 's/^/      /'
      return 1
    fi
    n=$((n + 1))
    sleep 1
  done

  printf "  \033[33m!\033[0m %-11s no probe answer on :%s after %ss — see %s\n" \
    "$name" "$(svc_port "$name")" "$tries" "$LOG_DIR/$name.log"
  return 1
}

# ── Dependency health check ───────────────────────────────────────────────────
check_deps() {
  echo "==> Checking infrastructure (postgres / redis / nats / influx)..."
  local down=0 label host port need d ok
  for d in "${DEPS[@]}"; do
    label="$(echo "$d" | cut -d: -f1)"
    host="$(echo "$d" | cut -d: -f2)"
    port="$(echo "$d" | cut -d: -f3)"
    need="$(echo "$d" | cut -d: -f4)"

    ok=1
    case "$label" in
      # pg_isready answers whether it will accept a connection, not just whether
      # something holds the port
      postgres) pg_isready -h "$host" -p "$port" >/dev/null 2>&1 || ok=0 ;;
      influx)   curl -sf -m 2 -o /dev/null "http://${host}:${port}/health" 2>/dev/null || ok=0 ;;
      *)        nc -z "$host" "$port" 2>/dev/null || ok=0 ;;
    esac

    if [ "$ok" -eq 1 ]; then
      printf "  \033[32m✓\033[0m %-9s %s:%s\n" "$label" "$host" "$port"
    elif [ "$need" = "optional" ]; then
      printf "  \033[33m!\033[0m %-9s %s:%s  (down, optional — charts and tick history)\n" \
        "$label" "$host" "$port"
    else
      printf "  \033[31m✗\033[0m %-9s %s:%s  (DOWN)\n" "$label" "$host" "$port"
      down=$((down + 1))
    fi
  done

  if [ "$down" -ne 0 ]; then
    echo ""
    echo "  $down required service(s) down. Start them with either:"
    echo "      brew services start postgresql@16 redis nats-server"
    echo "      cd hst-server && make up          # the docker stack"
    echo "  On other ports, set PG_PORT / REDIS_PORT / NATS_PORT."
    return 1
  fi
  echo "  Infrastructure up."
  return 0
}

# ── Tick feed liveness ────────────────────────────────────────────────────────
# Whether prices are actually moving, which is the first thing to check when the
# platform is up but nothing trades. Uses the nats CLI when it is installed and
# the monitoring endpoint when it is not, so neither is a hard requirement.
feed_status() {
  echo "==> Tick feed:"

  if command -v nats >/dev/null 2>&1; then
    local n
    n="$(timeout 4 nats sub 'hstquote.tick.*' 2>/dev/null | grep -c Received)" || true
    if [ "${n:-0}" -gt 0 ]; then
      printf "  \033[32m✓\033[0m %s ticks in 4s, the feed is live\n" "$n"
    else
      printf "  \033[31m✗\033[0m no ticks in 4s, the feed is dead\n"
    fi
    return 0
  fi

  local url="http://${NATS_HOST}:${NATS_MON_PORT}/varz" a b rate
  a="$(curl -sf -m 2 "$url" 2>/dev/null | sed -n 's/.*"in_msgs"[[:space:]]*:[[:space:]]*\([0-9]*\).*/\1/p' | head -1)" || true
  if [ -z "$a" ]; then
    printf "  \033[33m!\033[0m cannot tell: no nats CLI and no monitoring endpoint on :%s\n" "$NATS_MON_PORT"
    printf "    install it with 'brew install nats-io/nats-tools/nats', or start nats with -m %s\n" "$NATS_MON_PORT"
    return 0
  fi

  sleep 2
  b="$(curl -sf -m 2 "$url" 2>/dev/null | sed -n 's/.*"in_msgs"[[:space:]]*:[[:space:]]*\([0-9]*\).*/\1/p' | head -1)" || true
  b="${b:-$a}"
  rate=$(( (b - a) / 2 ))

  if [ "$rate" -gt 0 ]; then
    printf "  \033[32m✓\033[0m ~%s nats msgs/s, the bus is busy\n" "$rate"
  else
    printf "  \033[31m✗\033[0m no nats traffic in 2s, nothing is publishing\n"
  fi
}

# ── Lifecycle commands ────────────────────────────────────────────────────────
start_one() {
  local full
  full="$(full_name "$1")" || { echo "unknown service: $1"; echo "run: $0 list"; return 1; }
  free_port "$(svc_port "$full")"
  start_service "$full" $(health_port_overrides "$full")
  wait_service "$full" || true
}

stop_one() {
  local full pidfile
  full="$(full_name "$1")" || { echo "unknown service: $1"; echo "run: $0 list"; return 1; }
  pidfile="$LOG_DIR/$full.pid"
  echo "==> Stopping $full..."

  if [ -f "$pidfile" ]; then
    # SIGTERM, because every service drains in-flight work on it
    kill -TERM "$(cat "$pidfile")" 2>/dev/null || true
    local n=0
    while [ "$n" -lt 10 ] && is_running "$full"; do n=$((n + 1)); sleep 1; done
    is_running "$full" && kill -9 "$(cat "$pidfile")" 2>/dev/null || true
    rm -f "$pidfile"
  fi

  # scoped to this checkout's binaries, so an unrelated go process is left alone
  pkill -f "$BIN_DIR/$full" 2>/dev/null || true
  free_port "$(svc_port "$full")"
}

stop_all() {
  echo "==> Stopping all services..."
  # reverse order, so the engine and the feeds go before the api they report to
  local i name
  for (( i=${#SERVICES[@]}-1 ; i>=0 ; i-- )); do
    name="${SERVICES[$i]}"
    is_running "$name" || [ -f "$LOG_DIR/$name.pid" ] || continue
    stop_one "$name" >/dev/null 2>&1 || true
    echo "    stopped $name"
  done
  # anything left holding a known port, from a run whose pidfile was lost
  for name in "${SERVICES[@]}"; do
    pkill -f "$BIN_DIR/$name" 2>/dev/null || true
    free_port "$(svc_port "$name")"
  done
}

status() {
  check_deps || true
  echo ""
  echo "==> Services:"
  local name port pid
  for name in "${SERVICES[@]}"; do
    port="$(svc_port "$name")"
    if is_running "$name"; then
      pid="$(cat "$LOG_DIR/$name.pid")"
      if probe_ok "$name"; then
        printf "  \033[32m●\033[0m %-11s :%-6s pid %-7s healthy\n" "$name" "$port" "$pid"
      else
        printf "  \033[33m●\033[0m %-11s :%-6s pid %-7s no probe answer\n" "$name" "$port" "$pid"
      fi
    elif probe_ok "$name"; then
      printf "  \033[33m●\033[0m %-11s :%-6s (up, no pidfile)\n" "$name" "$port"
    else
      printf "  \033[31m○\033[0m %-11s :%-6s stopped\n" "$name" "$port"
    fi
  done
  echo ""
  feed_status
}

# health reports what each probe actually says, for when status is not enough.
health() {
  echo "==> Probe endpoints:"
  local name port path code
  for name in "${SERVICES[@]}"; do
    port="$(svc_port "$name")"
    path="$(svc_probe "$name")"
    code="$(curl -s -m 3 -o /dev/null -w '%{http_code}' "http://localhost:${port}${path}" 2>/dev/null)" || code="000"
    printf "  %-11s http://localhost:%-6s%-32s %s\n" "$name" "$port" "$path" "$code"

    # readiness is the one that checks dependencies, so it is worth its own line
    if [ "$name" != "hst-server" ]; then
      code="$(curl -s -m 3 -o /dev/null -w '%{http_code}' "http://localhost:${port}/readyz" 2>/dev/null)" || code="000"
      printf "  %-11s http://localhost:%-6s%-32s %s\n" "" "$port" "/readyz" "$code"
    fi
  done
}

tail_logs() {
  if [ -n "${1:-}" ]; then
    local full; full="$(full_name "$1")" || { echo "unknown service: $1"; return 1; }
    [ -f "$LOG_DIR/$full.log" ] || { echo "no log yet: $LOG_DIR/$full.log"; return 1; }
    tail -f "$LOG_DIR/$full.log"
  else
    # an unexpanded glob would make tail -f fail on a literal *.log
    local logs=("$LOG_DIR"/*.log)
    [ -e "${logs[0]}" ] || { echo "no logs yet in $LOG_DIR/"; return 1; }
    tail -f "${logs[@]}"
  fi
}

list_services() {
  echo "Infrastructure:"
  local d
  for d in "${DEPS[@]}"; do
    printf "  %-9s %s:%-6s %s\n" \
      "$(echo "$d" | cut -d: -f1)" "$(echo "$d" | cut -d: -f2)" \
      "$(echo "$d" | cut -d: -f3)" "$(echo "$d" | cut -d: -f4)"
  done
  echo ""
  echo "Services (aliases in parens):"
  echo "  hst-server  (server, api)    :$(svc_port hst-server)  http api, migrations, seeds, websockets"
  echo "  hst-core    (core, engine)   :$(svc_port hst-core)  trading engine, probe port"
  echo "  hst-quote   (quote, quotes)  :$(svc_port hst-quote)  FIX quote ingestion, probe port"
  echo "  hst-news    (news)           :$(svc_port hst-news)  RSS news ingestion, probe port"
}

summary() {
  local failed="${1:-0}"
  echo ""
  if [ "$failed" -eq 0 ]; then
    echo "All services up. Logs in: $LOG_DIR/"
  else
    echo "$failed service(s) did not come up. Logs in: $LOG_DIR/"
  fi
  echo ""
  echo "  api:        http://localhost:$(svc_port hst-server)"
  echo "  swagger:    http://localhost:$(svc_port hst-server)/swagger/admin/index.html"
  echo "              http://localhost:$(svc_port hst-server)/swagger/trader/index.html"
  echo ""
  echo "  status:   $0 status"
  echo "  health:   $0 health"
  echo "  tail all: $0 logs"
  echo "  stop all: $0 stop"
}

# Full bring-up.
up() {
  local do_build=1 failed=0 name
  [ "${1:-}" = "--no-build" ] && do_build=0

  if ! check_deps; then
    echo "==> Aborting: bring infrastructure up first."
    exit 1
  fi
  echo ""

  warn_health_port_clash
  stop_all
  echo ""

  [ "$do_build" -eq 1 ] && build_all

  # hst-server first and on its own: it migrates and seeds, and the ingestion
  # services read their configuration from it over /internal/v1
  start_service hst-server
  wait_service hst-server || { summary 1; exit 1; }
  echo ""

  for name in hst-core hst-quote hst-news; do
    start_service "$name" $(health_port_overrides "$name")
  done

  for name in hst-core hst-quote hst-news; do
    wait_service "$name" || failed=$((failed + 1))
  done

  summary "$failed"
  [ "$failed" -eq 0 ] || return 1
}

usage() {
  cat <<EOF
Usage: $0 <command> [args]

Commands:
  up [--no-build]    Check infra, stop all, build, start every service in order (default).
  build [svc]        Build every service into .service-bin, or just one.
  start <svc>        Start a single service (accepts alias, e.g. core or hst-core).
  stop [svc]         Stop one service, or all of them if no name is given.
  restart <svc>      Stop then start a single service.
  check | deps       Health-check postgres / redis / nats / influx.
  status | ps        Infra, every service, and whether prices are moving.
  health             The HTTP status each probe endpoint returns.
  logs [svc]         Tail all logs, or one service's log.
  list               List infrastructure and services with their ports.
  help               Show this help.

Ports come from each service's own .env; these only affect the health checks:
  PG_HOST PG_PORT           (localhost 5432)
  REDIS_HOST REDIS_PORT     (localhost 6379)
  NATS_HOST NATS_PORT       (localhost 4222)
  NATS_MON_PORT             (8222, for the tick-feed check without the nats CLI)
  INFLUX_HOST INFLUX_PORT   (localhost 8086)
  READY_TIMEOUT             (40 seconds per service)
  HST_NEWS_HEALTH_PORT      (8083, because hst-core already holds 8082)

Examples:
  $0                       # full bring-up
  $0 up --no-build         # bring up without rebuilding
  $0 check                 # is postgres/redis/nats/influx up?
  $0 restart core          # restart just the trading engine
  $0 logs quote            # tail hst-quote's log
  $0 stop                  # stop everything
EOF
}

# ── Dispatch ──────────────────────────────────────────────────────────────────
cmd="${1:-up}"; shift || true
case "$cmd" in
  up|all)   up "$@" ;;
  build)
    if [ -n "${1:-}" ]; then
      full="$(full_name "$1")" || { echo "unknown service: $1"; exit 1; }
      echo "==> Building $full..."; build_one "$full"
    else
      build_all
    fi ;;
  start)    [ -n "${1:-}" ] || { echo "usage: $0 start <svc>"; exit 1; }; check_deps || true; start_one "$1" ;;
  stop)     if [ -n "${1:-}" ]; then stop_one "$1"; else stop_all; echo "Stopped."; fi ;;
  restart)  [ -n "${1:-}" ] || { echo "usage: $0 restart <svc>"; exit 1; }; stop_one "$1"; sleep 1; start_one "$1" ;;
  check|deps|check-deps) check_deps ;;
  status|ps) status ;;
  health|probes) health ;;
  logs|log|tail) tail_logs "${1:-}" ;;
  list|services) list_services ;;
  help|-h|--help) usage ;;
  *) echo "unknown command: $cmd"; echo ""; usage; exit 1 ;;
esac

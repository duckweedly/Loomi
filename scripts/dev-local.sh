#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
API_HOST="${LOOMI_HOST:-http://127.0.0.1:18080}"
HTTP_ADDR="${HTTP_ADDR:-127.0.0.1:18080}"
DATABASE_URL="${DATABASE_URL:-postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable}"
WORKSPACE="${LOOMI_WORKSPACE:-$ROOT}"
PROVIDER="${LOOMI_PROVIDER:-local_codex}"
MODEL="${LOOMI_MODEL:-}"
WEB_PORT="${WEB_PORT:-5180}"
LOG_DIR="${LOOMI_DEV_LOG_DIR:-$ROOT/tmp/dev-local}"
RUN_SMOKE="${RUN_SMOKE:-0}"

mkdir -p "$LOG_DIR"

api_pid=""
web_pid=""

process_is_running() {
  local pid="$1"
  [[ -n "$pid" ]] && jobs -pr | grep -qx "$pid"
}

stop_children() {
  if [[ -n "$web_pid" ]] && kill -0 "$web_pid" 2>/dev/null; then
    kill "$web_pid" 2>/dev/null || true
  fi
  if [[ -n "$api_pid" ]] && kill -0 "$api_pid" 2>/dev/null; then
    kill "$api_pid" 2>/dev/null || true
  fi
}

trap stop_children EXIT INT TERM

cd "$ROOT"

if command -v migrate >/dev/null 2>&1; then
  migrate -path migrations -database "$DATABASE_URL" up
else
  printf 'migrate CLI not found; install migrate or apply migrations before starting local release helper.\n' >&2
  printf 'example: migrate -path migrations -database "$DATABASE_URL" up\n' >&2
  exit 1
fi

APP_ENV="${APP_ENV:-local}" \
HTTP_ADDR="$HTTP_ADDR" \
DATABASE_URL="$DATABASE_URL" \
LOOMI_WORKER_QUEUE_ENABLED="${LOOMI_WORKER_QUEUE_ENABLED:-true}" \
LOOMI_WORKER_QUEUE_PAUSED="${LOOMI_WORKER_QUEUE_PAUSED:-false}" \
go run ./cmd/loomi-api >"$LOG_DIR/api.log" 2>&1 &
api_pid="$!"

printf 'api starting: %s (log %s)\n' "$API_HOST" "$LOG_DIR/api.log"
doctor_ready=0
for _ in {1..60}; do
  if ! process_is_running "$api_pid"; then
    wait "$api_pid" 2>/dev/null || true
    printf 'api process exited before readiness; not starting web.\n' >&2
    printf 'api log: %s\n' "$LOG_DIR/api.log" >&2
    exit 1
  fi
  if go run ./cmd/loomi doctor --host "$API_HOST" --provider "$PROVIDER" >"$LOG_DIR/doctor.log" 2>&1; then
    doctor_ready=1
    break
  fi
  sleep 1
done
if [[ "$doctor_ready" != "1" ]]; then
  printf 'doctor did not report ready within 60s; not starting web.\n' >&2
  printf 'doctor log: %s\n' "$LOG_DIR/doctor.log" >&2
  printf 'api log: %s\n' "$LOG_DIR/api.log" >&2
  exit 1
fi

VITE_LOOMI_API_BASE_URL="$API_HOST" \
VITE_LOOMI_API_TOKEN="${VITE_LOOMI_API_TOKEN:-${LOOMI_API_TOKEN:-}}" \
bun run --cwd web dev --host 127.0.0.1 --port "$WEB_PORT" >"$LOG_DIR/web.log" 2>&1 &
web_pid="$!"

/bin/sleep 1
if ! process_is_running "$web_pid"; then
  wait "$web_pid" 2>/dev/null || true
  printf 'web process exited after startup; not running desktop doctor or smoke.\n' >&2
  printf 'web log: %s\n' "$LOG_DIR/web.log" >&2
  exit 1
fi
printf 'web starting: http://127.0.0.1:%s (log %s)\n' "$WEB_PORT" "$LOG_DIR/web.log"

if ! desktop_output="$(go run ./cmd/loomi doctor --host "$API_HOST" --provider "$PROVIDER" --desktop 2>&1)"; then
  printf '%s\n' "$desktop_output" >&2
  exit 1
fi
printf '%s\n' "$desktop_output"

if ! process_is_running "$web_pid"; then
  wait "$web_pid" 2>/dev/null || true
  printf 'web process exited after startup; not running smoke.\n' >&2
  printf 'web log: %s\n' "$LOG_DIR/web.log" >&2
  exit 1
fi

printf '\nselect/verify workspace:\n'
printf '  go run ./cmd/loomi smoke agent --host %q --provider %q --workspace %q --auto-approve --timeout 4m --prompt %q\n' "$API_HOST" "$PROVIDER" "$WORKSPACE" "Read AGENTS.md with workspace.read, then reply with local startup smoke complete."

if [[ "$RUN_SMOKE" == "1" ]]; then
  smoke_args=(smoke agent --host "$API_HOST" --provider "$PROVIDER" --workspace "$WORKSPACE" --auto-approve --timeout 4m --prompt "Read AGENTS.md with workspace.read, then reply with local startup smoke complete.")
  if [[ -n "$MODEL" ]]; then
    smoke_args+=(--model "$MODEL")
  fi
  if ! smoke_output="$(go run ./cmd/loomi "${smoke_args[@]}" 2>&1)"; then
    printf '%s\n' "$smoke_output" >&2
    exit 1
  fi
  printf '%s\n' "$smoke_output"
fi

printf '\nPress Ctrl-C to stop the API and web dev servers. Logs stay under %s.\n' "$LOG_DIR"
wait

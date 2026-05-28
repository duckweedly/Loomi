#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  exit 1
}

assert_contains() {
  local file="$1"
  local needle="$2"
  if ! grep -Fq "$needle" "$file"; then
    printf -- '--- %s ---\n' "$file" >&2
    cat "$file" >&2 || true
    fail "expected $file to contain: $needle"
  fi
}

assert_not_contains() {
  local file="$1"
  local needle="$2"
  if [[ -f "$file" ]] && grep -Fq "$needle" "$file"; then
    printf -- '--- %s ---\n' "$file" >&2
    cat "$file" >&2 || true
    fail "expected $file not to contain: $needle"
  fi
}

make_fake_bin() {
  local dir="$1"
  mkdir -p "$dir"
  cat >"$dir/go" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf 'go %s\n' "$*" >>"${DEV_LOCAL_TEST_CALLS:?}"
if [[ "$1 $2" == "run ./cmd/loomi-api" ]]; then
  if [[ "${DEV_LOCAL_API_EXIT:-0}" == "1" ]]; then
    printf 'api exited early\n'
    exit 42
  fi
  while true; do /bin/sleep 1; done
  exit 0
fi
if [[ "$1 $2 $3" == "run ./cmd/loomi doctor" ]]; then
  if [[ "$*" == *"--desktop"* && "${DEV_LOCAL_DESKTOP_DOCTOR:-success}" == "fail" ]]; then
    printf 'desktop doctor failed\n'
    exit 8
  fi
  if [[ "${DEV_LOCAL_DOCTOR:-success}" == "fail" ]]; then
    printf 'doctor failed\n'
    exit 7
  fi
  printf 'doctor ok\n'
  exit 0
fi
if [[ "$1 $2 $3" == "run ./cmd/loomi smoke" ]]; then
  if [[ "${DEV_LOCAL_SMOKE:-success}" == "fail" ]]; then
    printf 'smoke failed\n'
    exit 9
  fi
  printf 'smoke ok\n'
  exit 0
fi
exit 0
EOF
  chmod +x "$dir/go"
cat >"$dir/bun" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf 'bun %s\n' "$*" >>"${DEV_LOCAL_TEST_CALLS:?}"
if [[ "${DEV_LOCAL_WEB_EXIT:-0}" == "1" ]]; then
  printf 'web exited early\n'
  exit 43
fi
while true; do /bin/sleep 1; done
exit 0
EOF
  chmod +x "$dir/bun"
  cat >"$dir/sleep" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  chmod +x "$dir/sleep"
}

run_helper() {
  local dir="$1"
  shift
  (
    cd "$ROOT"
    env \
      PATH="$dir/bin:/usr/bin:/bin" \
      DEV_LOCAL_TEST_CALLS="$dir/calls.log" \
      LOOMI_DEV_LOG_DIR="$dir/logs" \
      "$@" \
      bash scripts/dev-local.sh
  ) >"$dir/stdout.log" 2>"$dir/stderr.log"
}

run_start_helper() {
  local dir="$1"
  shift
  (
    cd "$ROOT"
    env \
      PATH="$dir/bin:/usr/bin:/bin" \
      DEV_LOCAL_TEST_CALLS="$dir/calls.log" \
      LOOMI_START_LOG_DIR="$dir/start-logs" \
      "$@" \
      bash scripts/start-local.sh
  ) >"$dir/stdout.log" 2>"$dir/stderr.log"
}

test_missing_migrate_fails_before_startup() {
  local dir
  dir="$(mktemp -d)"
  make_fake_bin "$dir/bin"
  if run_helper "$dir"; then
    fail "missing migrate exited 0"
  fi
  assert_contains "$dir/stderr.log" "migrate CLI not found"
  assert_not_contains "$dir/calls.log" "go run ./cmd/loomi-api"
  assert_not_contains "$dir/calls.log" "bun run --cwd web dev"
}

test_api_exit_blocks_web_startup() {
  local dir
  dir="$(mktemp -d)"
  make_fake_bin "$dir/bin"
  cat >"$dir/bin/migrate" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  chmod +x "$dir/bin/migrate"
  if run_helper "$dir" DEV_LOCAL_API_EXIT=1 DEV_LOCAL_DOCTOR=fail; then
    fail "api early exit returned 0"
  fi
  assert_contains "$dir/stderr.log" "api process exited before readiness"
  assert_contains "$dir/stderr.log" "$dir/logs/api.log"
  assert_not_contains "$dir/calls.log" "bun run --cwd web dev"
}

test_doctor_timeout_blocks_web_startup() {
  local dir
  dir="$(mktemp -d)"
  make_fake_bin "$dir/bin"
  cat >"$dir/bin/migrate" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  chmod +x "$dir/bin/migrate"
  if run_helper "$dir" DEV_LOCAL_DOCTOR=fail; then
    fail "doctor timeout exited 0"
  fi
  assert_contains "$dir/stderr.log" "doctor did not report ready"
  assert_contains "$dir/stderr.log" "$dir/logs/doctor.log"
  assert_contains "$dir/stderr.log" "$dir/logs/api.log"
  assert_not_contains "$dir/calls.log" "bun run --cwd web dev"
}

test_web_exit_blocks_smoke_startup() {
  local dir
  dir="$(mktemp -d)"
  make_fake_bin "$dir/bin"
  cat >"$dir/bin/migrate" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  chmod +x "$dir/bin/migrate"
  if run_helper "$dir" DEV_LOCAL_WEB_EXIT=1 RUN_SMOKE=1 DEV_LOCAL_SMOKE=fail; then
    fail "web early exit returned 0"
  fi
  assert_contains "$dir/stderr.log" "web process exited after startup"
  assert_contains "$dir/stderr.log" "$dir/logs/web.log"
  assert_not_contains "$dir/calls.log" "go run ./cmd/loomi smoke"
}

test_desktop_doctor_failure_exits_nonzero() {
  local dir
  dir="$(mktemp -d)"
  make_fake_bin "$dir/bin"
  cat >"$dir/bin/migrate" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  chmod +x "$dir/bin/migrate"
  if run_helper "$dir" DEV_LOCAL_DESKTOP_DOCTOR=fail; then
    fail "desktop doctor failure exited 0"
  fi
  assert_contains "$dir/stdout.log" "web starting:"
  assert_contains "$dir/stderr.log" "desktop doctor failed"
}

test_smoke_failure_exits_nonzero() {
  local dir
  dir="$(mktemp -d)"
  make_fake_bin "$dir/bin"
  cat >"$dir/bin/migrate" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  chmod +x "$dir/bin/migrate"
  if run_helper "$dir" RUN_SMOKE=1 DEV_LOCAL_SMOKE=fail; then
    fail "smoke failure exited 0"
  fi
  assert_contains "$dir/stderr.log" "smoke failed"
}

test_start_local_wrapper_uses_user_log_dir() {
  local dir
  dir="$(mktemp -d)"
  make_fake_bin "$dir/bin"
  cat >"$dir/bin/migrate" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  chmod +x "$dir/bin/migrate"
  if run_start_helper "$dir" DEV_LOCAL_DOCTOR=fail; then
    fail "start-local doctor failure exited 0"
  fi
  assert_contains "$dir/stderr.log" "doctor did not report ready"
  assert_contains "$dir/stderr.log" "$dir/start-logs/doctor.log"
  assert_contains "$dir/stderr.log" "$dir/start-logs/api.log"
}

test_missing_migrate_fails_before_startup
test_api_exit_blocks_web_startup
test_doctor_timeout_blocks_web_startup
test_web_exit_blocks_smoke_startup
test_desktop_doctor_failure_exits_nonzero
test_smoke_failure_exits_nonzero
test_start_local_wrapper_uses_user_log_dir

printf 'scripts/dev-local.test.sh: ok\n'

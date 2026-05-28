#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export LOOMI_DEV_LOG_DIR="${LOOMI_START_LOG_DIR:-${LOOMI_LOG_DIR:-$ROOT/tmp/start-local}}"

exec "$ROOT/scripts/dev-local.sh" "$@"

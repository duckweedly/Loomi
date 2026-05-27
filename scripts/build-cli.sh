#!/usr/bin/env sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
OUTPUT=${OUTPUT:-"$ROOT/dist/loomi"}

if [ -z "${VERSION:-}" ]; then
  if git -C "$ROOT" describe --tags --always --dirty >/dev/null 2>&1; then
    VERSION=$(git -C "$ROOT" describe --tags --always --dirty)
  else
    VERSION=dev
  fi
fi

if [ -z "${COMMIT:-}" ]; then
  if git -C "$ROOT" rev-parse --short HEAD >/dev/null 2>&1; then
    COMMIT=$(git -C "$ROOT" rev-parse --short HEAD)
  else
    COMMIT=unknown
  fi
fi

DATE=${DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}

mkdir -p "$(dirname "$OUTPUT")"

cd "$ROOT"
go build -trimpath \
  -ldflags "-s -w -X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" \
  -o "$OUTPUT" \
  ./cmd/loomi

"$OUTPUT" version

#!/usr/bin/env sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
PREFIX=${PREFIX:-"$HOME/.local"}
TARGET=${TARGET:-"$PREFIX/bin/loomi"}

"$ROOT/scripts/build-cli.sh"

if [ -e "$TARGET" ] && [ "${LOOMI_INSTALL_OVERWRITE:-}" != "1" ]; then
  echo "target exists: $TARGET" >&2
  echo "set LOOMI_INSTALL_OVERWRITE=1 to replace it" >&2
  exit 1
fi

mkdir -p "$(dirname "$TARGET")"
install -m 0755 "$ROOT/dist/loomi" "$TARGET"

"$TARGET" version
echo "installed $TARGET"

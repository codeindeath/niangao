#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_BIN="${GO_BIN:-go}"

if ! command -v "$GO_BIN" >/dev/null 2>&1; then
  if [ -x "$HOME/.local/toolchains/go1.26.3/bin/go" ]; then
    GO_BIN="$HOME/.local/toolchains/go1.26.3/bin/go"
  else
    echo "go executable not found. Set GO_BIN or install Go 1.23+." >&2
    exit 127
  fi
fi

cd "$ROOT_DIR/backend"
"$GO_BIN" test ./...

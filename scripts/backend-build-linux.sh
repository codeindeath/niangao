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

OUT="${1:-$ROOT_DIR/backend/niangao-backend-linux-amd64}"

cd "$ROOT_DIR/backend"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 "$GO_BIN" build -o "$OUT" ./cmd/server
file "$OUT"

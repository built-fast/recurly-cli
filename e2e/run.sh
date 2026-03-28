#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_ROOT"

# --- Dependency checks ---
missing=()
for cmd in bats npx jq curl; do
  command -v "$cmd" >/dev/null 2>&1 || missing+=("$cmd")
done
if [ ${#missing[@]} -gt 0 ]; then
  echo "ERROR: Missing required dependencies: ${missing[*]}" >&2
  echo "Please install them before running the e2e test suite." >&2
  exit 1
fi

# --- Build CLI binary ---
echo "==> Building CLI binary..."
make build
RECURLY_BINARY="$PROJECT_ROOT/recurly"

# --- Select a random available port ---
get_random_port() {
  python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1",0)); print(s.getsockname()[1]); s.close()'
}
PRISM_PORT=$(get_random_port)
PRISM_URL="http://127.0.0.1:${PRISM_PORT}"

# --- Cleanup trap ---
PRISM_PID=""
cleanup() {
  if [ -n "$PRISM_PID" ]; then
    kill "$PRISM_PID" 2>/dev/null || true
    wait "$PRISM_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

# --- Start Prism mock server ---
echo "==> Starting Prism mock server on port ${PRISM_PORT}..."
npx @stoplight/prism-cli mock openapi/api.yaml \
  --port "$PRISM_PORT" \
  --host 127.0.0.1 \
  --dynamic \
  > /dev/null 2>&1 &
PRISM_PID=$!

# --- Wait for Prism to be ready ---
echo "==> Waiting for Prism to be ready..."
elapsed=0
timeout=30
while [ "$elapsed" -lt "$timeout" ]; do
  if curl -so /dev/null "${PRISM_URL}/" 2>&1; then
    echo "==> Prism is ready."
    break
  fi
  sleep 1
  elapsed=$((elapsed + 1))
done

if [ "$elapsed" -ge "$timeout" ]; then
  echo "ERROR: Prism did not become ready within ${timeout} seconds." >&2
  exit 1
fi

# --- Run BATS tests ---
echo "==> Running e2e tests..."
export PRISM_URL
export RECURLY_BINARY

bats e2e/*.bats

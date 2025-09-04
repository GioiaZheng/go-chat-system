#!/usr/bin/env bash
set -e

cd "$(dirname "$0")"

echo "[1/3] Killing old webapi (if any)..."
pids=$(pgrep -f "./webapi" || true)
if [ -n "$pids" ]; then
  echo "Killing: $pids"
  kill -9 $pids || true
else
  echo "No running webapi found."
fi

echo "[2/3] Building..."
go build -o webapi ./cmd/webapi

echo "[3/3] Running (foreground)..."
./webapi

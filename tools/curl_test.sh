#!/usr/bin/env bash
# Generic curl test runner for the API.
# Configuration is via env vars; no absolute URLs hardcoded.
#
# Env:
#   SCHEME=http|https        (default: http)
#   HOST=example.com         (default: localhost)
#   PORT=3000                (default: 3000)
#   API_PREFIX=/api/v1       (default: /api/v1)
#   TOKEN=<bearer>           (optional; if set, Authorization header will be sent)

set -euo pipefail

SCHEME="${SCHEME:-http}"
HOST="${HOST:-localhost}"
PORT="${PORT:-3000}"
API_PREFIX="${API_PREFIX:-/api/v1}"
BASE="${BASE:-${SCHEME}://${HOST}:${PORT}}"

# Print command to stderr (so command substitution captures only stdout)
run() {
  >&2 printf '+ '
  >&2 printf '%q ' "$@"
  >&2 echo
  "$@"
}

auth_args=()
if [[ -n "${TOKEN:-}" ]]; then
  auth_args=(-H "Authorization: Bearer ${TOKEN}")
fi

# Public: liveness
run curl -sS "${BASE}${API_PREFIX}/liveness"

# Auth examples (uncomment and adapt as needed):
# run curl -sS "${auth_args[@]}" -H "Content-Type: application/json" \
#   -X POST "${BASE}${API_PREFIX}/messages" \
#   -d '{"content":"hi","toUserId":"USER_ID"}'

#!/usr/bin/env bash
# Generic curl test runner for the API.
# This script runs a minimal smoke test through major endpoints.
#
# Env vars (override with `VAR=value ./test.sh`):
#   SCHEME=http|https        (default: http)
#   HOST=example.com         (default: localhost)
#   PORT=3000                (default: 3000)
#   API_PREFIX=/api/v1       (default: /api/v1)

set -euo pipefail

SCHEME="${SCHEME:-http}"
HOST="${HOST:-localhost}"
PORT="${PORT:-3000}"
API_PREFIX="${API_PREFIX:-/api/v1}"
BASE="${BASE:-${SCHEME}://${HOST}:${PORT}}"

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

echo "== Health check =="
run curl -sS "${BASE}${API_PREFIX}/liveness" | jq .

echo
echo "== Register user =="
REGISTER=$(run curl -sS -X POST "${BASE}${API_PREFIX}/register" \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com","password":"pass","gender":"female"}' | tee /dev/stderr)
TOKEN=$(echo "$REGISTER" | jq -r '.data.token')
USER_ID=$(echo "$REGISTER" | jq -r '.data.user.id')

auth_args=(-H "Authorization: Bearer ${TOKEN}")

echo
echo "== Get current user =="
run curl -sS "${auth_args[@]}" "${BASE}${API_PREFIX}/users/me" | jq .

echo
echo "== Search users (self excluded) =="
run curl -sS "${auth_args[@]}" "${BASE}${API_PREFIX}/users/search?q=ali" | jq .

echo
echo "== Create group =="
GROUP=$(run curl -sS -X POST "${BASE}${API_PREFIX}/groups" \
  "${auth_args[@]}" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"testgroup\",\"member_ids\":[\"${USER_ID}\"]}" | tee /dev/stderr)
GROUP_ID=$(echo "$GROUP" | jq -r '.data.group.id')

echo
echo "== Send message =="
MSG=$(run curl -sS -X POST "${BASE}${API_PREFIX}/messages" \
  "${auth_args[@]}" \
  -H "Content-Type: application/json" \
  -d "{\"conversation_id\":\"${GROUP_ID}\",\"content\":\"Hello group!\"}" | tee /dev/stderr)
MSG_ID=$(echo "$MSG" | jq -r '.data.resource.id')

echo
echo "== Comment message =="
run curl -sS -X POST "${BASE}${API_PREFIX}/messages/${MSG_ID}/comment" \
  "${auth_args[@]}" \
  -H "Content-Type: application/json" \
  -d '{"type":"text","content":"nice!"}' | jq .

echo
echo "== Get comments =="
run curl -sS "${BASE}${API_PREFIX}/messages/${MSG_ID}/comment" "${auth_args[@]}" | jq .

echo
echo "== Uncomment (remove last comment) =="
run curl -sS -X POST "${BASE}${API_PREFIX}/messages/${MSG_ID}/uncomment" "${auth_args[@]}" | jq .

echo
echo "== Delete message =="
run curl -sS -X DELETE "${BASE}${API_PREFIX}/messages/${MSG_ID}" "${auth_args[@]}" | jq .

echo
echo "== Done ✅ =="

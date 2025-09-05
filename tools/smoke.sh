#!/usr/bin/env bash
# Purpose: quick smoke test for OpenAPI routes.
# Notes:
# - We do NOT 'set -e' so that even if one request fails, the script continues.
# - A status code of 000 means the HTTP request failed (connection refused, etc).

set -u

BASE="${BASE:-http://localhost:3000}"   # e.g., http://localhost:3000
PFX="${PFX:-/api/v1}"                   # e.g., /api/v1

probe() {
  local method="$1" path="$2" data="${3:-}"
  local code
  if [[ -n "${data}" ]]; then
    code=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" \
      -H "Content-Type: application/json" \
      "${BASE}${PFX}${path}" -d "$data" || echo "000")
  else
    code=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" \
      "${BASE}${PFX}${path}" || echo "000")
  fi
  printf "%-6s %-38s -> %s\n" "$method" "${PFX}${path}" "$code"
}

echo "Public:"
probe GET  /liveness
echo

echo "Protected (expect 401 without token):"
probe POST /conversations '{"name":"smoke","memberIds":[]}'
probe GET  /conversations
probe PUT  /users/set_username '{"username":"alice"}'
probe PUT  /users/set_photo '{"avatarUrl":"x"}'
probe GET  /users/me
probe GET  "/users/search?q=a"
probe GET  /users/profile/USER_ID
probe POST /groups '{"name":"g"}'
probe GET  /groups
probe GET  /groups/GROUP_ID
probe PUT  /groups/GROUP_ID/name '{"name":"g2"}'
probe PUT  /groups/GROUP_ID/photo '{"avatarUrl":"x"}'
probe POST /groups/GROUP_ID/members '{"userIds":["u"]}'
probe DELETE /groups/GROUP_ID/members
probe GET  "/messages?chat_type=private&target_id=U"
probe POST /messages '{"chat_type":"private","target_id":"U","content":"hi"}'
probe GET  /messages/MSG_ID
probe DELETE /messages/MSG_ID
probe POST /messages/MSG_ID/forward '{"target_type":"group","target_id":"G"}'
probe GET  /messages/MSG_ID/comment
probe POST /messages/MSG_ID/comment '{"content":"nice"}'
probe POST /messages/MSG_ID/uncomment

echo
echo "Compat (should exist, may return 401):"
probe GET "/messages-group?target_id=G"
probe GET "/messages-private?target_id=U"

echo
echo "Hint: '000' means connection failed. Make sure the server is running:"
echo "  go run ./cmd/webapi"

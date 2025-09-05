cat > tools/smoke.sh <<'EOF'
#!/usr/bin/env bash
# Unauthenticated smoke test (no hardcoded absolute URLs).
# Configure either:
#   BASE="http://<host>:<port>"  and optionally PFX (default: /api/v1)
# or:
#   SCHEME=http|https  HOST=<host>  PORT=<port>  and optionally PFX
#
# Example:
#   BASE="http://localhost:3000" ./tools/smoke.sh
#   SCHEME=http HOST=localhost PORT=3000 ./tools/smoke.sh

set -euo pipefail

BASE="${BASE:-}"
PFX="${PFX:-/api/v1}"

if [[ -z "$BASE" ]]; then
  if [[ -n "${SCHEME:-}" && -n "${HOST:-}" && -n "${PORT:-}" ]]; then
    BASE="${SCHEME}://${HOST}:${PORT}"
  else
    echo "ERROR: Set BASE or SCHEME/HOST/PORT (no hardcoded defaults in this script)." >&2
    exit 2
  fi
fi

probe() {
  local method="$1" path="$2"
  local code
  code=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" "${BASE}${path}")
  printf "%-6s %-36s -> %s\n" "$method" "${path}" "$code"
}

echo "Public:"
probe GET  "${PFX}/liveness"

echo
echo "Protected (expect 401 without token):"
probe POST "${PFX}/conversations"
probe GET  "${PFX}/conversations"
probe PUT  "${PFX}/users/set_username"
probe PUT  "${PFX}/users/set_photo"
probe GET  "${PFX}/users/me"
probe GET  "${PFX}/users/search?q=a"
probe GET  "${PFX}/users/profile/USER_ID"
probe POST "${PFX}/groups"
probe GET  "${PFX}/groups"
probe GET  "${PFX}/groups/GROUP_ID"
probe PUT  "${PFX}/groups/GROUP_ID/name"
probe PUT  "${PFX}/groups/GROUP_ID/photo"
probe POST "${PFX}/groups/GROUP_ID/members"
probe DELETE "${PFX}/groups/GROUP_ID/members"
probe GET  "${PFX}/messages?chat_type=private&target_id=U"
probe POST "${PFX}/messages"
probe GET  "${PFX}/messages/MSG_ID"
probe DELETE "${PFX}/messages/MSG_ID"
probe POST "${PFX}/messages/MSG_ID/forward"
probe GET  "${PFX}/messages/MSG_ID/comment"
probe POST "${PFX}/messages/MSG_ID/comment"
probe POST "${PFX}/messages/MSG_ID/uncomment"

echo
echo "Compat (should exist, may return 401):"
probe GET  "${PFX}-group?target_id=G"
probe GET  "${PFX}-private?target_id=U"

echo
echo "Hint: '000' means connection failed. Make sure the server is running:"
echo "  go run ./cmd/webapi"
EOF

chmod +x tools/smoke.sh

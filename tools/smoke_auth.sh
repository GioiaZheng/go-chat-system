cat > tools/smoke_auth.sh <<'EOF'
#!/usr/bin/env bash
# Authenticated smoke test (no hardcoded absolute URLs).
# Configure either:
#   BASE="http://<host>:<port>" and optionally PFX (default: /api/v1)
# or:
#   SCHEME=http|https HOST=<host> PORT=<port> and optionally PFX
# Authenticated smoke test (no hardcoded absolute URLs).
# Configure either:
#   BASE="http://<host>:<port>" and optionally PFX (default: /api/v1)
# or:
#   SCHEME=http|https HOST=<host> PORT=<port> and optionally PFX
#
# Login expects: {"name":"...", "password":"..."}
# Token extracted from: data[0].token

set -euo pipefail

BASE="${BASE:-}"
BASE="${BASE:-}"
PFX="${PFX:-/api/v1}"
NAME="${NAME:-alice}"
NAME="${NAME:-alice}"
PASSWORD="${PASSWORD:-passw0rd}"

if [[ -z "$BASE" ]]; then
  if [[ -n "${SCHEME:-}" && -n "${HOST:-}" && -n "${PORT:-}" ]]; then
    BASE="${SCHEME}://${HOST}:${PORT}"
  else
    echo "ERROR: Set BASE or SCHEME/HOST/PORT (no hardcoded defaults in this script)." >&2
    exit 2
  fi
fi

run() { >&2 printf '+ '; >&2 printf '%q ' "$@"; >&2 echo; "$@"; }

echo "== Register (best-effort) =="
code=$(run curl -s -o /dev/null -w "%{http_code}" \
  -X POST "${BASE}${PFX}/register" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${NAME}\",\"password\":\"${PASSWORD}\"}")
echo "${code}"

echo
echo "== Login =="
LOGIN_JSON=$(run curl -sS \
  -X POST "${BASE}${PFX}/session" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${NAME}\",\"password\":\"${PASSWORD}\"}")

if command -v jq >/dev/null 2>&1; then
  TOKEN="$(printf '%s' "$LOGIN_JSON" | jq -r '.data[0].token // empty')"
else
  TOKEN="$(printf '%s' "$LOGIN_JSON" \
    | grep -oE '"token"[[:space:]]*:[[:space:]]*"[^"]+"' \
    | head -n1 \
    | sed -E 's/.*:[[:space:]]*"([^"]+)".*/\1/')"
fi

if [[ -z "${TOKEN:-}" || "${TOKEN}" == "null" ]]; then
  echo "Login response did not contain a usable token."
  echo "Response was: $LOGIN_JSON"
  exit 1
fi

echo "Token obtained."
AUTH=(-H "Authorization: Bearer ${TOKEN}")

probe() {
  local method="$1" path="$2" data="${3:-}"
  local code
  if [[ -n "$data" ]]; then
    code=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" \
      "${AUTH[@]}" -H "Content-Type: application/json" \
      "${BASE}${path}" -d "$data")
      "${BASE}${path}" -d "$data")
  else
    code=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" \
      "${AUTH[@]}" "${BASE}${path}")
      "${AUTH[@]}" "${BASE}${path}")
  fi
  printf "%-6s %-36s -> %s\n" "$method" "${path}" "$code"
  printf "%-6s %-36s -> %s\n" "$method" "${path}" "$code"
}

echo
echo "== Protected endpoints with token (expect 200/201) =="
probe GET  "${PFX}/users/me"
probe GET  "${PFX}/users/search?q=a"
probe POST "${PFX}/conversations" "{\"name\":\"test\",\"memberIds\":[\"${NAME}\"]}"
probe GET  "${PFX}/conversations"
probe POST "${PFX}/groups" "{\"name\":\"demo-group\"}"
probe GET  "${PFX}/groups"
probe GET  "${PFX}/messages?chat_type=private&target_id=${NAME}"
probe POST "${PFX}/messages" "{\"chat_type\":\"private\",\"target_id\":\"${TOKEN}\",\"content\":\"hi\"}"
probe GET  "${PFX}/users/me"
probe GET  "${PFX}/users/search?q=a"
probe POST "${PFX}/conversations" "{\"name\":\"test\",\"memberIds\":[\"${NAME}\"]}"
probe GET  "${PFX}/conversations"
probe POST "${PFX}/groups" "{\"name\":\"demo-group\"}"
probe GET  "${PFX}/groups"
probe GET  "${PFX}/messages?chat_type=private&target_id=${NAME}"
probe POST "${PFX}/messages" "{\"chat_type\":\"private\",\"target_id\":\"${TOKEN}\",\"content\":\"hi\"}"

echo
echo "Done."
EOF

chmod +x tools/smoke_auth.sh

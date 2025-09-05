#!/usr/bin/env bash
# Authenticated smoke test:
# - Try register (best-effort)
# - Login to obtain a Bearer token
# - Call a subset of protected endpoints with the token
#
# Your backend expects login JSON to use: {"name":"...","password":"..."}  (NOT "username").
# The login response contains token at: data[0].token
#
# Env:
#   BASE=http://localhost:3000   (default)
#   PFX=/api/v1                  (default)
#   NAME=alice                   (default)   # <-- note: NAME
#   PASSWORD=passw0rd            (default)

set -euo pipefail

BASE="${BASE:-http://localhost:3000}"
PFX="${PFX:-/api/v1}"
NAME="${NAME:-alice}"             # <-- your LoginRequest uses "name"
PASSWORD="${PASSWORD:-passw0rd}"

# Print command to STDERR, run command and keep STDOUT clean for capture.
run() {
  >&2 printf '+ '
  >&2 printf '%q ' "$@"
  >&2 echo
  "$@"
}

# ---- 1) Register (best-effort) ----
# If your register expects {"name": "...", "password": "..."} it will work.
# If it expects extra fields, add them here.
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

# Extract token at data[0].token; jq is preferred; grep/sed fallback if jq is missing.
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
      "${BASE}${PFX}${path}" -d "$data")
  else
    code=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" \
      "${AUTH[@]}" "${BASE}${PFX}${path}")
  fi
  printf "%-6s %-36s -> %s\n" "$method" "${PFX}${path}" "$code"
}

echo
echo "== Protected endpoints with token (expect 200/201) =="
probe GET  /users/me
probe GET  "/users/search?q=a"
probe POST /conversations "{\"name\":\"test\",\"memberIds\":[\"${NAME}\"]}"
probe GET  /conversations
probe POST /groups "{\"name\":\"demo-group\"}"
probe GET  /groups
probe GET  "/messages?chat_type=private&target_id=${NAME}"
probe POST /messages "{\"chat_type\":\"private\",\"target_id\":\"${TOKEN}\",\"content\":\"hi\"}"

echo
echo "Done."

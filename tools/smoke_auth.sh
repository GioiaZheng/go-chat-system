#!/usr/bin/env bash
set -euo pipefail

BASE="${BASE:-}"
NAME="${NAME:-alice}"

if [[ -z "$BASE" ]]; then
  if [[ -n "${SCHEME:-}" && -n "${HOST:-}" && -n "${PORT:-}" ]]; then
    BASE="${SCHEME}://${HOST}:${PORT}"
  else
    echo "ERROR: Set BASE or SCHEME/HOST/PORT." >&2
    exit 2
  fi
fi

run(){ >&2 printf '+ '; >&2 printf '%q ' "$@"; >&2 echo; "$@"; }
pp(){ if command -v jq >/dev/null 2>&1; then jq .; else cat; fi; }
json_get(){
  local js="$1" jqexpr="$2" re="$3" out
  if command -v jq >/dev/null 2>&1; then
    out="$(printf '%s' "$js" | jq -r "$jqexpr // empty")"
  else
    out="$(printf '%s' "$js" | grep -oE "$re" | head -n1 | sed -E 's/.*:[[:space:]]*"([^"]+)".*/\1/')"
  fi
  printf '%s' "$out"
}

echo "== Liveness =="
run curl -sS "${BASE}${PFX}/liveness" | pp

echo
echo "== Login (simplified) =="
LOGIN_JSON=$(run curl -sS \
  -X POST "${BASE}${PFX}/session" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${NAME}\"}" | tee /dev/stderr)

TOKEN="$(json_get "$LOGIN_JSON" '.data.token' '"token"[[:space:]]*:[[:space:]]*"[^"]+"')"
USER_ID="$(json_get "$LOGIN_JSON" '.data.user.id' '"id"[[:space:]]*:[[:space:]]*"[^"]+"')"
[[ -z "$TOKEN" || "$TOKEN" == "null" ]] && { echo "Login failed. See above." >&2; exit 1; }
[[ -z "$USER_ID" || "$USER_ID" == "null" ]] && { echo "No user id. See above." >&2; exit 1; }
echo "Token: ${TOKEN:0:8}...  UserID: $USER_ID"
AUTH=(-H "Authorization: Bearer ${TOKEN}")

probe(){
  local method="$1" path="$2" data="${3:-}"
  local code
  if [[ -n "$data" ]]; then
    code=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" \
      "${AUTH[@]}" -H "Content-Type: application/json" \
      "${BASE}${path}" -d "$data")
  else
    code=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" \
      "${AUTH[@]}" "${BASE}${path}")
  fi
  printf "%-6s %-48s -> %s\n" "$method" "$path" "$code"
}

echo
echo "== Basic protected =="
probe GET  "${PFX}/users/me"
probe GET  "${PFX}/users/search?q=${NAME:0:1}"

# --- 可选：建群/会话（如果你要保留也行） ---
echo
echo "== Create conversation (optional reliable path) =="
CONV_JSON=$(run curl -sS -X POST "${BASE}${PFX}/conversations" \
  "${AUTH[@]}" -H "Content-Type: application/json" \
  -d "{\"name\":\"demo-conv\",\"member_ids\":[\"${USER_ID}\"]}" | tee /dev/stderr) || true
CONV_ID="$(json_get "$CONV_JSON" '.data.conversation.id' '"id"[[:space:]]*:[[:space:]]*"[^"]+"')"
[[ -n "${CONV_ID:-}" ]] && echo "CONVERSATION_ID=${CONV_ID}"

# ------------------ 发送消息（带自动回退） ------------------
echo
echo "== Send message =="
MSG_JSON=""
MSG_ID=""

if [[ -n "${CONV_ID:-}" ]]; then
  echo "-- Try new payload: conversation_id --"
  MSG_JSON=$(run curl -sS -X POST "${BASE}${PFX}/messages" \
    "${AUTH[@]}" -H "Content-Type: application/json" \
    -d "{\"conversation_id\":\"${CONV_ID}\",\"content\":\"hello group\"}" | tee /dev/stderr)
  # 如果成功，会有 data.resource.id
  MSG_ID="$(json_get "$MSG_JSON" '.data.resource.id' '"id"[[:space:]]*:[[:space:]]*"[^"]+"')"
fi

if [[ -z "$MSG_ID" ]]; then
  echo "-- Fallback: legacy payload chat_type+target_id (send to self as private) --"
  MSG_JSON=$(run curl -sS -X POST "${BASE}${PFX}/messages" \
    "${AUTH[@]}" -H "Content-Type: application/json" \
    -d "{\"chat_type\":\"private\",\"target_id\":\"${USER_ID}\",\"content\":\"hello self\"}" | tee /dev/stderr)
  MSG_ID="$(json_get "$MSG_JSON" '.data.resource.id' '"id"[[:space:]]*:[[:space:]]*"[^"]+"')"
fi

if [[ -z "$MSG_ID" ]]; then
  echo "ERROR: cannot extract message id. See responses above." >&2
  exit 1
fi
echo "MSG_ID=${MSG_ID}"

# ------------------ 拉取消息（带自动回退） ------------------
echo
echo "== Pull messages =="
if [[ -n "${CONV_ID:-}" ]]; then
  echo "-- Try new query: conversation_id --"
  CODE=$(curl -s -o /dev/null -w "%{http_code}" "${BASE}${PFX}/messages?conversation_id=${CONV_ID}" "${AUTH[@]}")
  if [[ "$CODE" == "200" ]]; then
    run curl -sS "${BASE}${PFX}/messages?conversation_id=${CONV_ID}" "${AUTH[@]}" | pp
  else
    echo "-- Fallback: legacy query chat_type+target_id (self) --"
    run curl -sS "${BASE}${PFX}/messages?chat_type=private&target_id=${USER_ID}" "${AUTH[@]}" | pp
  fi
else
  echo "-- Legacy query chat_type+target_id (self) --"
  run curl -sS "${BASE}${PFX}/messages?chat_type=private&target_id=${USER_ID}" "${AUTH[@]}" | pp
fi

# ------------------ 评论 / 取消评论 / 删除 ------------------
echo
echo "== Comment / Uncomment / Delete =="
run curl -sS -X POST "${BASE}${PFX}/messages/${MSG_ID}/comment" \
  "${AUTH[@]}" -H "Content-Type: application/json" \
  -d '{"type":"text","content":"nice!"}' | pp

run curl -sS "${BASE}${PFX}/messages/${MSG_ID}/comment" "${AUTH[@]}" | pp
run curl -sS -X POST "${BASE}${PFX}/messages/${MSG_ID}/uncomment" "${AUTH[@]}" | pp
run curl -sS -X DELETE "${BASE}${PFX}/messages/${MSG_ID}" "${AUTH[@]}" | pp

echo
echo "Done ✅"

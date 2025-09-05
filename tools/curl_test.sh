#!/usr/bin/env bash
# Robust curl test script that never stops on first error.
# It prints raw responses + HTTP status codes for easy debugging.

set -u  # keep -u to catch typos; avoid -e/-o pipefail so we don't exit early

BASE="http://localhost:3000/api/v1"

# Generate unique usernames to avoid collisions with existing accounts
SUFFIX=$(date +%s)
ALICE_NAME="alice_${SUFFIX}"
BOB_NAME="bob_${SUFFIX}"

hr() { printf '\n------ %s ------\n' "$*"; }

REQUEST() {
  local label="$1" method="$2" url="$3" body="${4:-}" token="${5:-}"

  hr "$label"
  if [[ -n "$body" ]]; then
    echo "[curl] $method $url"
    echo "[body] $body"
    curl -sS -D /tmp/ct_headers -o /tmp/ct_body \
      -X "$method" "$url" \
      -H "Content-Type: application/json" \
      ${token:+ -H "Authorization: Bearer $token"} \
      --data "$body" || true
  else
    echo "[curl] $method $url"
    curl -sS -D /tmp/ct_headers -o /tmp/ct_body \
      -X "$method" "$url" \
      ${token:+ -H "Authorization: Bearer $token"} \
      || true
  fi

  status=$(sed -n '1s/.* \([0-9][0-9][0-9]\).*/\1/p' /tmp/ct_headers)
  echo "[status] ${status:-unknown}"
  echo "[headers]"
  sed 's/^/  /' /tmp/ct_headers
  echo "[body]"
  cat /tmp/ct_body
  echo

  LAST_BODY="$(cat /tmp/ct_body)"
}

# Simple regex-based extractors (good enough for our payloads)
field_token() {
  sed -n 's/.*"token":"\([^"]*\)".*/\1/p' <<<"$LAST_BODY"
}

# Prefer new shape: data.group.id ; fallback to old: groupId
field_group_id() {
  # new shape
  local v
  v=$(sed -n 's/.*"group":{"id":"\([^"]*\)".*/\1/p' <<<"$LAST_BODY")
  if [[ -n "$v" ]]; then echo "$v"; return; fi
  # fallback old
  sed -n 's/.*"groupId":"\([^"]*\)".*/\1/p' <<<"$LAST_BODY"
}

########################################
# Flow
########################################

REQUEST "Register user ${ALICE_NAME}" "POST" "$BASE/register" \
  "{\"name\":\"${ALICE_NAME}\",\"password\":\"pwd\"}"
ALICE_TOKEN="$(field_token)"
echo "ALICE_TOKEN=$ALICE_TOKEN"

REQUEST "Register user ${BOB_NAME}" "POST" "$BASE/register" \
  "{\"name\":\"${BOB_NAME}\",\"password\":\"pwd\"}"
BOB_TOKEN="$(field_token)"
echo "BOB_TOKEN=$BOB_TOKEN"

REQUEST "GET /users/me as ${ALICE_NAME}" "GET" "$BASE/users/me" "" "$ALICE_TOKEN"

REQUEST "Search users (query=${ALICE_NAME}) as alice" "GET" \
  "$BASE/users/search?q=${ALICE_NAME}" "" "$ALICE_TOKEN"

REQUEST "Create group 'lab-group' with members [${BOB_NAME}] as alice" "POST" \
  "$BASE/groups" \
  "{\"name\":\"lab-group\",\"members\":[\"${BOB_TOKEN}\"]}" \
  "$ALICE_TOKEN"
GROUP_ID="$(field_group_id)"
echo "GROUP_ID=$GROUP_ID"

REQUEST "GET /groups (list) as alice" "GET" "$BASE/groups" "" "$ALICE_TOKEN"

if [[ -n "$GROUP_ID" ]]; then
  REQUEST "GET /groups/:id as alice" "GET" "$BASE/groups/${GROUP_ID}" "" "$ALICE_TOKEN"
else
  hr "Skip GET /groups/:id because GROUP_ID is empty"
fi

REQUEST "POST /messages (private: ${ALICE_NAME} -> ${BOB_NAME})" "POST" \
  "$BASE/messages" \
  "{\"content\":\"hello bob (from ${ALICE_NAME})\",\"toUserId\":\"${BOB_TOKEN}\"}" \
  "$ALICE_TOKEN"

REQUEST "GET /messages?chat_type=private&target_id=${BOB_NAME} (as alice)" "GET" \
  "$BASE/messages?chat_type=private&target_id=${BOB_TOKEN}" "" "$ALICE_TOKEN"

if [[ -n "$GROUP_ID" ]]; then
  REQUEST "POST /messages (group: ${ALICE_NAME} -> lab-group)" "POST" \
    "$BASE/messages" \
    "{\"content\":\"hello group\",\"toGroupId\":\"${GROUP_ID}\"}" \
    "$ALICE_TOKEN"

  REQUEST "GET /messages?chat_type=group&target_id=$GROUP_ID (as alice)" "GET" \
    "$BASE/messages?chat_type=group&target_id=${GROUP_ID}" "" "$ALICE_TOKEN"
else
  hr "Skip group message tests because GROUP_ID is empty"
fi

hr "Done."

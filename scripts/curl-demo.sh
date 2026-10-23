#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
DEMO_TMP="$(mktemp -d)"
trap 'rm -rf "$DEMO_TMP"' EXIT

LAST_BODY="$DEMO_TMP/body"
LAST_HEADERS="$DEMO_TMP/headers"

request() {
  local title="$1"
  local expected="$2"
  shift 2

  local status
  status="$(curl -sS -D "$LAST_HEADERS" -o "$LAST_BODY" -w '%{http_code}' "$@")"
  if [[ "$status" != "$expected" ]]; then
    echo "FAIL: $title: status=$status, want=$expected" >&2
    cat "$LAST_BODY" >&2
    exit 1
  fi
  echo "OK: $title -> $status" >&2
}

register_and_login() {
  local email="$1"

  request "register $email" 201 -X POST "$BASE_URL/register" \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"$email\",\"password\":\"secret123\"}"
  if grep -Eqi 'password|hash' "$LAST_BODY"; then
    echo 'FAIL: register response exposes password data' >&2
    exit 1
  fi

  request "login $email" 200 -X POST "$BASE_URL/login" \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"$email\",\"password\":\"secret123\"}"
  python3 -c 'import json,sys; print(json.load(sys.stdin)["token"])' <"$LAST_BODY"
}

suffix="${RANDOM}-$$"
alice_email="alice-${suffix}@example.com"
bob_email="bob-${suffix}@example.com"

request 'health' 200 "$BASE_URL/health"
request 'protected endpoint without token' 401 "$BASE_URL/notes"
if ! tr -d '\r' <"$LAST_HEADERS" | grep -qi '^WWW-Authenticate: Bearer$'; then
  echo 'FAIL: 401 response has no WWW-Authenticate: Bearer' >&2
  exit 1
fi

alice_token="$(register_and_login "$alice_email")"
bob_token="$(register_and_login "$bob_email")"

request 'Alice creates Note' 201 -X POST "$BASE_URL/notes" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $alice_token" \
  -d '{"title":"Private note","text":"Alice owns it"}'
note_id="$(python3 -c 'import json,sys; print(json.load(sys.stdin)["id"])' <"$LAST_BODY")"

request 'Bob cannot read Alice Note' 403 "$BASE_URL/notes/$note_id" \
  -H "Authorization: Bearer $bob_token"

request 'Bob list contains only Bob Notes' 200 "$BASE_URL/notes" \
  -H "Authorization: Bearer $bob_token"
python3 -c 'import json,sys; assert json.load(sys.stdin) == []' <"$LAST_BODY"

request 'Alice reads her Note' 200 "$BASE_URL/notes/$note_id" \
  -H "Authorization: Bearer $alice_token"

request 'Alice updates her Note' 200 -X PUT "$BASE_URL/notes/$note_id" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $alice_token" \
  -d '{"title":"Updated","text":"Ownership checked in UseCase"}'

request 'Alice deletes her Note' 204 -X DELETE "$BASE_URL/notes/$note_id" \
  -H "Authorization: Bearer $alice_token"
if [[ -s "$LAST_BODY" ]]; then
  echo 'FAIL: 204 response must have an empty body' >&2
  exit 1
fi

request 'deleted Note is missing' 404 "$BASE_URL/notes/$note_id" \
  -H "Authorization: Bearer $alice_token"

echo 'curl demo: PASS'

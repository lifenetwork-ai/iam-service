#!/bin/bash
# E2E Update Identifier Flow Test Script (Full Matrix)
# Runs register → verify → update-identifier → verify for four scenarios.
# Writes a linear log suitable for pasting into Markdown.

set -euo pipefail

API_URL="http://localhost:8080"
TENANT_ID="ea815eb6-bb85-49ad-bf1a-1a71862f4c7a"
WEBHOOK_TOKEN="963e0036-282b-4c84-9fa4-d57186f4b142"
WEBHOOK_URL="https://webhook.site/token/${WEBHOOK_TOKEN}/requests?page=1&sorting=newest"
RESULTS_FILE="e2e_update_identifier_results.txt"

require_bin() {
  command -v "$1" >/dev/null 2>&1 || { echo "Missing binary: $1"; exit 1; }
}
require_bin curl
require_bin jq
require_bin grep
require_bin sed

print_section() {
  {
    echo
    echo "=============================="
    echo "$1"
    echo "=============================="
    echo
  } | tee -a "$RESULTS_FILE"
}

log() {
  echo -e "$1" | tee -a "$RESULTS_FILE"
}

# Extract a 6-digit code from arbitrary text (prefer *123456*; otherwise any 6 digits)
_extract_code() {
  local body="$1"
  local code
  code=$(printf "%s" "$body" | grep -oE '\*[0-9]{6}\*' | grep -oE '[0-9]{6}' | head -n1 || true)
  if [[ -z "${code:-}" ]]; then
    code=$(printf "%s" "$body" | grep -oE '[^0-9]([0-9]{6})[^0-9]' | grep -oE '[0-9]{6}' | head -n1 || true)
  fi
  printf "%s" "${code:-}"
}

# Poll webhook.site for a new OTP. If receiver_hint is provided, prefer entries containing it.
# Stops when the newest observed code differs from the initial snapshot or after timeout.
get_otp() {
  local receiver="$1"
  local max_tries=30   # ~60s if sleep=2
  local sleep_s=2

  # Get the latest OTP at start (snapshot)
  local old
  old=$(curl -s "$WEBHOOK_URL" \
    | jq -r --arg r "$receiver" '
        .data[]
        | .content
        | select(. != null and . != "")
        | fromjson
        | select(.to==$r)
        | .message
      ' \
    | grep -oE '[0-9]{6}' \
    | head -n1)

  # Poll for a new code that differs from snapshot
  for ((i=1; i<=max_tries; i++)); do
    sleep "$sleep_s"

    local code
    code=$(curl -s "$WEBHOOK_URL" \
      | jq -r --arg r "$receiver" '
          .data[]
          | .content
          | select(. != null and . != "")
          | fromjson
          | select(.to==$r)
          | .message
        ' \
      | grep -oE '[0-9]{6}' \
      | head -n1)

    if [[ -n "$code" && "$code" != "$old" ]]; then
      echo "$code"
      return 0
    fi
  done

  # Fall back to snapshot if no new OTP arrives
  echo "$old"
  return 1
}

# Privatix Temp Mail via RapidAPI: Generate random email and fetch OTP
RAPIDAPI_KEY="b18f15e3aamsh30b0b383e8b8759p17dafbjsn19863ac35017"
RAPIDAPI_HOST="privatix-temp-mail-v1.p.rapidapi.com"

get_domain() {
  curl -s --request GET \
    --url "https://$RAPIDAPI_HOST/request/domains/" \
    --header "x-rapidapi-host: $RAPIDAPI_HOST" \
    --header "x-rapidapi-key: $RAPIDAPI_KEY" \
    | jq -r '.[0]'
}

create_disposable_email() {
  local login domain email
  login="testuser$RANDOM"
  domain=$(get_domain)
  email="$login$domain"
  echo "$email $login $domain"
}

md5_hash() {
  if command -v md5sum >/dev/null 2>&1; then
    echo -n "$1" | md5sum | awk '{print $1}'
  else
    echo -n "$1" | md5 | awk '{print $4}'
  fi
}

# Poll Privatix inbox for OTP
get_email_otp() {
  local email_md5="$1"
  local max_tries=10
  local sleep_s=3
  local otp=""
  for ((i=1; i<=max_tries; i++)); do
    sleep "$sleep_s"
    local inbox=$(curl -s --request GET \
      --url "https://$RAPIDAPI_HOST/request/mail/id/$email_md5/" \
      --header "x-rapidapi-host: $RAPIDAPI_HOST" \
      --header "x-rapidapi-key: $RAPIDAPI_KEY")
    otp=$(echo "$inbox" | jq -e 'if type == "array" and length > 0 then .[0].mail_text else empty end' | grep -oE '[0-9]{6}' | head -n1)
    if [[ -z "$otp" ]]; then
      otp=$(echo "$inbox" | jq -e 'if type == "array" and length > 0 then .[0].mail_text_only else empty end' | grep -oE '[0-9]{6}' | head -n1)
    fi
    if [[ -z "$otp" ]]; then
      otp=$(echo "$inbox" | jq -e 'if type == "array" and length > 0 then .[0].mail_html else empty end' | grep -oE '[0-9]{6}' | head -n1)
    fi
    if [[ -n "$otp" ]]; then
      echo "$otp"
      return 0
    fi
    echo "No OTP yet, retrying... ($i/$max_tries)"
  done
  echo ""
  return 1
}


# Clean log
: > "$RESULTS_FILE"

# Robust curl wrapper returning raw JSON or failing fast with context
post_json() {
  local url="$1"
  local data="$2"
  shift 2
  curl -sS -X POST "$url" \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    "$@" \
    -d "$data"
}

get_auth() {
  local url="$1"
  local token="$2"
  shift 2 || true
  curl -sS -X GET "$url" \
    -H "accept: application/json" \
    -H "Authorization: Bearer $token" \
    "$@"
}

expect_ok() {
  local json="$1"
  local ctx="$2"
  local status
  status=$(printf "%s" "$json" | jq -r '.status // empty')
  if [[ "$status" != "200" && "$status" != 200 ]]; then
    echo "Failure at: $ctx"
    echo "$json"
    exit 1
  fi
}

register_with_phone() {
  local base_phone="$1"
  local lang="$2"
  local max_attempts=5
  local attempt=1
  local phone="$base_phone"

  while (( attempt <= max_attempts )); do
    REGISTER1=$(curl -s -X POST "$API_URL/api/v1/users/register" \
      -H "accept: application/json" \
      -H "X-Tenant-Id: $TENANT_ID" \
      -H "Content-Type: application/json" \
      -d "{\"phone\": \"$phone\", \"lang\": \"$lang\"}")

    local status=$(echo "$REGISTER1" | jq -r '.status')
    local code=$(echo "$REGISTER1" | jq -r '.code')
    local flow_id=$(echo "$REGISTER1" | jq -r '.data.verification_flow.flow_id')

    if [[ "$status" == "200" ]]; then
      echo "$phone $flow_id"
      return 0
    elif [[ "$code" == "MSG_IDENTIFIER_ALREADY_EXISTS" || "$code" == "MSG_RATE_LIMIT_EXCEEDED" ]]; then
      phone="+84$(( ${phone:3} + 1 ))"
      ((attempt++))
      sleep 2
    else
      >&2 echo "Registration failed: $REGISTER1"
      exit 1
    fi
  done

  >&2 echo "Failed to register after $max_attempts attempts."
  exit 1
}


# ========== Scenario 1: Phone → Email ==========
print_section "Scenario 1: Phone → Email"
read -r PHONE1 FLOW1 <<< "$(register_with_phone "+84321555566" "en")"
read -r EMAIL1 EMAIL_LOGIN EMAIL_DOMAIN <<< "$(create_disposable_email)"
EMAIL_MD5=$(md5_hash "$EMAIL1")

log "Get OTP for $PHONE1 from webhook"
OTP1=$(get_otp "$PHONE1")
log "OTP: $OTP1"

log "Verify registration with OTP"
VERIFY1=$(post_json "$API_URL/api/v1/users/challenge-verify" \
  "$(jq -cn --arg flow "$FLOW1" --arg code "$OTP1" '{flow_id:$flow,code:$code,type:"register"}')" \
  -H "X-Tenant-Id: $TENANT_ID")
log "$VERIFY1"
expect_ok "$VERIFY1" "Scenario1 verify register"
SESSION1=$(printf "%s" "$VERIFY1" | jq -r '.data.session_token // empty')

log "Initiate update to email: $EMAIL1"
UPDATE1=$(post_json "$API_URL/api/v1/users/me/update-identifier" \
  "$(jq -cn --arg id "$EMAIL1" '{new_identifier:$id,identifier_type:"email"}')" \
  -H "X-Tenant-Id: $TENANT_ID" -H "Authorization: Bearer $SESSION1")
log "$UPDATE1"
expect_ok "$UPDATE1" "Scenario1 initiate update email"
FLOW2=$(printf "%s" "$UPDATE1" | jq -r '.data.flow_id // empty')

log "Get OTP for $EMAIL1 from Privatix Temp Mail"
OTP2=$(get_email_otp "$EMAIL_MD5")
log "OTP: $OTP2"

log "Verify update identifier with OTP"
VERIFY2=$(post_json "$API_URL/api/v1/users/challenge-verify" \
  "$(jq -cn --arg flow "$FLOW2" --arg code "$OTP2" '{flow_id:$flow,code:$code,type:"register"}')" \
  -H "X-Tenant-Id: $TENANT_ID")
log "$VERIFY2"
expect_ok "$VERIFY2" "Scenario1 verify update email"
SESSION2=$(printf "%s" "$VERIFY2" | jq -r '.data.session_token // empty')

log "Login with new email: $EMAIL1"
LOGIN1=$(post_json "$API_URL/api/v1/users/challenge-with-email" \
  "$(jq -cn --arg email "$EMAIL1" '{email:$email}')" \
  -H "X-Tenant-Id: $TENANT_ID")
log "$LOGIN1"

log "Login with old phone (should fail): $PHONE1"
LOGIN2=$(post_json "$API_URL/api/v1/users/challenge-with-phone" \
  "$(jq -cn --arg phone "$PHONE1" '{phone:$phone}')" \
  -H "X-Tenant-Id: $TENANT_ID")
log "$LOGIN2"

log "Get user profile with new session token"
PROFILE1=$(get_auth "$API_URL/api/v1/users/me" "$SESSION2" -H "X-Tenant-Id: $TENANT_ID")
log "$PROFILE1"
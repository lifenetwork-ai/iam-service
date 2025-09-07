#!/bin/bash
set -euo pipefail

RAPIDAPI_KEY="b18f15e3aamsh30b0b383e8b8759p17dafbjsn19863ac35017"
RAPIDAPI_HOST="privatix-temp-mail-v1.p.rapidapi.com"

require_bin() {
  command -v "$1" >/dev/null 2>&1 || { echo "Missing binary: $1"; exit 1; }
}
require_bin curl
require_bin jq
require_bin md5sum || require_bin md5 # macOS uses md5

# Get a random domain from Privatix
get_domain() {
  curl -s --request GET \
    --url "https://$RAPIDAPI_HOST/request/domains/" \
    --header "x-rapidapi-host: $RAPIDAPI_HOST" \
    --header "x-rapidapi-key: $RAPIDAPI_KEY" \
    | jq -r '.[0]'
}

# Generate random email
create_disposable_email() {
  local login domain email
  login="testuser$RANDOM"
  domain=$(get_domain)
  email="$login$domain"
  echo "$email $login $domain"
}

# Compute MD5 hash (cross-platform)
md5_hash() {
  if command -v md5sum >/dev/null 2>&1; then
    echo -n "$1" | md5sum | awk '{print $1}'
  else
    echo -n "$1" | md5 | awk '{print $4}'
  fi
}

# Poll for OTP in inbox
get_email_otp() {
  local email_md5="$1"
  local max_tries=10
  local sleep_s=3
  local otp=""
  for ((i=1; i<=max_tries; i++)); do
    sleep "$sleep_s"
    local msgs=$(curl -s --request GET \
      --url "https://$RAPIDAPI_HOST/request/mail/id/$email_md5/" \
      --header "x-rapidapi-host: $RAPIDAPI_HOST" \
      --header "x-rapidapi-key: $RAPIDAPI_KEY")
    otp=$(echo "$msgs" | grep -oE '[0-9]{6}' | head -n1)
    if [[ -n "$otp" ]]; then
      echo "$otp"
      return 0
    fi
  done
  echo ""
  return 1
}

# Optionally delete a message by mail_id
# Usage: delete_message <mail_id>
delete_message() {
  local mail_id="$1"
  curl -s --request GET \
    --url "https://$RAPIDAPI_HOST/request/delete/id/$mail_id/" \
    --header "x-rapidapi-host: $RAPIDAPI_HOST" \
    --header "x-rapidapi-key: $RAPIDAPI_KEY"
}

read -r EMAIL1 LOGIN DOMAIN <<< "$(create_disposable_email)"
echo "Generated email: $EMAIL1"
EMAIL_MD5=$(md5_hash "$EMAIL1")
echo "MD5: $EMAIL_MD5"

# Print instructions for sending a test email
cat <<EOF
Send an email to: $EMAIL1
Subject/body should contain a 6-digit code (e.g. 123456)
EOF

OTP=$(get_email_otp "$EMAIL_MD5")
echo "Fetched OTP (should be empty unless you send an email): $OTP"

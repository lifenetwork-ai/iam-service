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

get_domain() {
  curl -s --request GET \
    --url "https://$RAPIDAPI_HOST/request/domains/" \
    --header "x-rapidapi-host: $RAPIDAPI_HOST" \
    --header "x-rapidapi-key: $RAPIDAPI_KEY" \
    | jq -r '.[0]'
}

# create_disposable_email() {
#   local login domain email
#   login="testuser$RANDOM"
#   domain=$(get_domain)
#   email="$login$domain"
#   echo "$email $login $domain"
# }

md5_hash() {
  if command -v md5sum >/dev/null 2>&1; then
    echo -n "$1" | md5sum | awk '{print $1}'
  else
    echo -n "$1" | md5 | awk '{print $4}'
  fi
}

get_email_otp() {
  local email_md5="aa4c6edb82f2ca3e2b559d7eb49bc04f"
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

read -r EMAIL1 LOGIN DOMAIN <<< "$(create_disposable_email)"
echo "Generated email: $EMAIL1"
EMAIL_MD5=$(md5_hash "$EMAIL1")
echo "MD5: $EMAIL_MD5"
echo "Send an email to: $EMAIL1 with a 6-digit OTP in the body (e.g. *123456*) and run this script again to test extraction."

OTP=$(get_email_otp "$EMAIL_MD5")
echo "Fetched OTP (should be empty unless you send an email): $OTP"

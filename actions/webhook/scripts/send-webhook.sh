#!/usr/bin/env bash
set -euo pipefail

: "${SHOGUN_WEBHOOK_URL:?SHOGUN_WEBHOOK_URL is required}"
: "${SHOGUN_SECRET:?SHOGUN_SECRET is required}"
: "${SHOGUN_PAYLOAD:?SHOGUN_PAYLOAD is required}"
: "${GITHUB_OUTPUT:?GITHUB_OUTPUT is required}"
# Covers an explicit empty curl-timeout from the action and standalone runs.
SHOGUN_TIMEOUT="${SHOGUN_TIMEOUT:-30}"

# HMAC-SHA256 over the exact bytes that curl will send, hex, lowercase.
signature=$(printf '%s' "$SHOGUN_PAYLOAD" \
  | openssl dgst -sha256 -hmac "$SHOGUN_SECRET" \
  | awk '{print $NF}')
signature=$(printf '%s' "$signature" | tr '[:upper:]' '[:lower:]')

body_file=$(mktemp)
trap 'rm -f "$body_file"' EXIT

# --data-raw sends the bytes as given and never treats a leading @ as a filename.
status=$(curl --silent --show-error \
  --max-time "$SHOGUN_TIMEOUT" \
  --output "$body_file" \
  --write-out '%{http_code}' \
  --request POST "$SHOGUN_WEBHOOK_URL" \
  --header 'Content-Type: application/json' \
  --header "X-Shogun-HMAC: sha256=$signature" \
  --data-raw "$SHOGUN_PAYLOAD") || true

response=$(cat "$body_file")

{
  printf 'status-code=%s\n' "$status"
  printf 'response-body<<SHOGUN_RESPONSE_EOF\n%s\nSHOGUN_RESPONSE_EOF\n' "$response"
} >> "$GITHUB_OUTPUT"

# 200 is the only success status this endpoint returns
if [ "$status" = "200" ]; then
  echo "Shogun accepted the webhook (HTTP $status): $response"
  exit 0
fi

case "$status" in
  000)
    echo "::error::Could not reach $SHOGUN_WEBHOOK_URL, network error or timeout after ${SHOGUN_TIMEOUT}s"
    ;;
  400)
    echo "::error::400 Bad Request - payload is not a flat JSON object of string values, or no auth header was sent: $response"
    ;;
  401)
    echo "::error::401 Unauthorized - X-Shogun-HMAC did not match HMAC-SHA256(body, secret), check inputs.secret: $response"
    ;;
  403)
    echo "::error::403 Forbidden - webhook is paused, resume it from the Shogun UI or PATCH /admin/hook/{slug}/resume: $response"
    ;;
  404)
    echo "::error::404 Not Found - no webhook for the slug in inputs.url (the registry is loaded from the DB at startup): $response"
    ;;
  500)
    echo "::error::500 Server Error - pipeline missing or not declaring the ci_webhook trigger: $response"
    ;;
  *)
    echo "::error::Webhook returned HTTP $status: $response"
    ;;
esac
exit 1

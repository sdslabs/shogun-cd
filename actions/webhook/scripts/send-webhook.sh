#!/usr/bin/env bash
set -euo pipefail

: "${SHOGUN_WEBHOOK_URL:?SHOGUN_WEBHOOK_URL is required}"
: "${SHOGUN_SECRET:?SHOGUN_SECRET is required}"
: "${SHOGUN_PAYLOAD:?SHOGUN_PAYLOAD is required}"
: "${GITHUB_OUTPUT:?GITHUB_OUTPUT is required}"
# Covers an explicit empty curl-timeout from the action and standalone runs.
SHOGUN_TIMEOUT="${SHOGUN_TIMEOUT:-30}"
if [[ ! "$SHOGUN_TIMEOUT" =~ ^[0-9]+([.][0-9]+)?$ ]] || [[ ! "$SHOGUN_TIMEOUT" =~ [1-9] ]]; then
  echo '::error::inputs.curl-timeout must be a positive number of seconds'
  exit 1
fi
case "$SHOGUN_WEBHOOK_URL" in
  http://*|https://*) ;;
  *)
    echo '::error::inputs.url must be an HTTP or HTTPS URL'
    exit 1
    ;;
esac

# HMAC-SHA256 over the exact bytes that curl will send, hex, lowercase.
signature=$(printf '%s' "$SHOGUN_PAYLOAD" \
  | openssl dgst -sha256 -hmac "$SHOGUN_SECRET" \
  | awk '{print $NF}')
signature=$(printf '%s' "$signature" | tr '[:upper:]' '[:lower:]')

body_file=$(mktemp)
trap 'rm -f "$body_file"' EXIT

# --data-raw sends the bytes as given and never treats a leading @ as a filename.
# Ignore user curl configuration so it cannot enable redirects or retries.
curl_exit=0
status=$(curl --disable --silent --show-error \
  --max-time "$SHOGUN_TIMEOUT" \
  --output "$body_file" \
  --write-out '%{http_code}' \
  --request POST "$SHOGUN_WEBHOOK_URL" \
  --header 'Content-Type: application/json' \
  --header "X-Shogun-HMAC: sha256=$signature" \
  --data-raw "$SHOGUN_PAYLOAD") || curl_exit=$?
status="${status:-000}"

response=$(<"$body_file")

# The response may come from a proxy, so treat it as arbitrary text. Use a fresh
# token that cannot terminate the output early or resume runner commands.
while :; do
  response_token="SHOGUN_$(openssl rand -hex 16)"
  [[ "$response" != *"$response_token"* ]] && break
done

{
  printf 'status-code=%s\n' "$status"
  printf 'response-body<<%s\n%s\n%s\n' "$response_token" "$response" "$response_token"
} >> "$GITHUB_OUTPUT"

if [ -n "$response" ]; then
  echo 'Webhook response:'
  printf '::stop-commands::%s\n' "$response_token"
  printf '%s\n' "$response"
  printf '::%s::\n' "$response_token"
fi

# HTTP 200 means the server accepted this trigger, even if reading the response
# failed afterward. Retrying here could start a duplicate deployment.
if [ "$status" = "200" ]; then
  if [ "$curl_exit" -ne 0 ]; then
    echo "::warning::Received HTTP 200, but curl could not finish reading the response (exit $curl_exit). The webhook is treated as accepted; check Shogun for the run."
  fi
  echo "Shogun accepted the webhook (HTTP $status)"
  exit 0
fi

case "$status" in
  000)
    echo "::error::No HTTP response received (curl exit $curl_exit). Check the URL, connectivity, TLS, and ${SHOGUN_TIMEOUT}s timeout. Check Shogun before retrying; the request may have reached the server."
    ;;
  400)
    echo '::error::400 Bad Request - payload must be a flat JSON object of string values'
    ;;
  401)
    echo '::error::401 Unauthorized - check inputs.secret'
    ;;
  403)
    echo '::error::403 Forbidden - webhook is paused, resume it from the Shogun UI'
    ;;
  404)
    echo '::error::404 Not Found - check the webhook URL and slug'
    ;;
  500)
    echo '::error::500 Server Error - check the response and Shogun server logs'
    ;;
  *)
    echo "::error::Webhook returned HTTP $status"
    ;;
esac
exit 1

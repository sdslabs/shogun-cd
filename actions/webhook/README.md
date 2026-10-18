# Shogun Webhook action

Trigger a Shogun-CD pipeline from another repository's GitHub Actions workflow.
The action converts the payload to JSON, signs those bytes using HMAC-SHA256, and sends one POST
request with an `X-Shogun-HMAC: sha256=<hex>` header.

Create an active webhook for a pipeline declaring `ci_webhook`, then save its
secret as the repository or organization secret `SHOGUN_HMAC_SECRET`. Use the
full webhook URL in the format `https://<host>/api/hook/<slug>`.

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      # Run this after your build and image push have succeeded.
      # Replace <commit-sha> with the full commit SHA of the action version to use.
      - uses: sdslabs/shogun-cd/actions/webhook@<commit-sha>
        id: deploy
        with:
          url: https://shogun.example.com/api/hook/<slug>
          secret: ${{ secrets.SHOGUN_HMAC_SECRET }}
          payload: |
            IMAGE_TAG: ${{ github.sha }}
            BRANCH: ${{ github.ref_name }}
```

No checkout or GitHub token permission is required by this action. The runner
must be able to reach your Shogun server. Use HTTPS outside trusted local networks.
Always use the public webhook URL in the format `https://<host>/api/hook/<slug>`.

## Inputs

| Input | Required | Default | Meaning |
| --- | --- | --- | --- |
| `url` | Yes | — | Full webhook URL in the format `https://<host>/api/hook/<slug>`. |
| `secret` | Yes | — | Webhook secret shown when the webhook was created. Pass it from GitHub secrets. |
| `payload` | No | Empty | One `KEY: value` pair per line, used as pipeline trigger values. Omit when no values are needed. |
| `curl-timeout` | No | `30` | Positive request timeout in seconds; decimals are supported. An empty string uses 30. |

Use `payload: |` for simple key-value lines, without `-` list markers. Keys must
start with a letter or underscore and contain only letters, digits, or underscores.
The first colon separates the key from the value, so URLs and image tags work.
Blank lines are ignored; surrounding whitespace is trimmed. Values such as `true`
and `123` remain strings. Duplicate keys are rejected.

This is a simple line format, not full YAML: quotes, `#`, and backslashes in values
are literal text. Do not add quotes around values. Each value must fit on one line;
raw JSON payloads and `-` lists are not accepted.

The action handles JSON encoding and escaping internally before signing and
sending. An omitted, empty, or whitespace-only `payload` sends an empty object.

## Outputs and success

- `status-code`: HTTP status, or `000` if no HTTP response was received.
- `response-body`: server response text, with trailing newlines trimmed. Error
  responses may be non-JSON, and transfer failures may leave an incomplete body.

HTTP 200 means Shogun accepted the trigger; it does not mean the deployment has
finished successfully. If curl fails after receiving HTTP 200, the action warns
and still treats the webhook as accepted. Other statuses fail the action.

Requests are never automatically retried and redirects are not followed. A timeout
can occur after the server accepted a request: check Shogun before manually retrying
to avoid duplicate deployments. The action ignores local `.curlrc` configuration.
Responses are printed as text with GitHub workflow-command processing disabled.

## Runner dependencies

The action requires Bash, curl, OpenSSL, jq, awk, tr, mktemp, and rm. It is intended
for Ubuntu GitHub-hosted runners and Linux self-hosted runners. On runners with
`apt-get` or `apk`, missing tools are installed automatically using root or
passwordless sudo. Command names are mapped to their packages (`gawk` and
`coreutils` where needed). Bash must already be installed to start the action.

On other runners, provide these tools beforehand. macOS and Windows runners
have not been verified.

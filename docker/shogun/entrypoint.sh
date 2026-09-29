#!/bin/sh

set -eu
CONFIG="/etc/shogun/config.yaml"
CONTAINER_DATA_DIR="/var/lib/shogun"

log() { printf '[config-builder] %s\n' "$*"; }
die() { printf '[config-builder] ERROR: %s\n' "$*" >&2; exit 1; }

need() {
    eval "value=\${$1:-}"
    [ -n "$value" ] || die "required environment variable $1 is not set"
    printf '%s' "$value"
}

need_int() {
    _v="$(need "$1")"
    case "$_v" in
        *[!0-9]*) die "$1 must be a number (got '$_v')" ;;
    esac
    printf '%s' "$_v"
}

sed_escape() { printf '%s' "$1" | sed -e 's/[\\&|]/\\&/g'; }

yaml_str() { printf "'%s'" "$(printf '%s' "$1" | sed -e "s/'/''/g")"; }

#reordering the sample.config.yaml will break this script as it depends on the current ordering of it
S_TOP='1,/^git:/'
S_GIT='/^git:/,/^api:/'
S_API='/^api:/,/^database:/'
S_DB='/^database:/,$'

set_str() { sed -i "$1 s|^\( *$2:\).*|\1 $(sed_escape "$(yaml_str "$3")")|" "$CONFIG"; }
set_raw() { sed -i "$1 s|^\( *$2:\).*|\1 $(sed_escape "$3")|" "$CONFIG"; }

if [ "${SHOGUN_SKIP_CONFIG_RENDER:-false}" = "true" ]; then
    log "SHOGUN_SKIP_CONFIG_RENDER=true — leaving $CONFIG untouched"
    exec "$@"
fi

[ -f "$CONFIG" ] || die "no config template at $CONFIG"
[ -w "$CONFIG" ] || die "$CONFIG is not writable (bind-mounted read-only?)"

for _section in git api database; do
    grep -q "^${_section}:" "$CONFIG" \
        || die "$CONFIG has no '${_section}:' section — not a shogun config template?"
done

log "rendering $CONFIG from environment"

set_raw "$S_TOP" debug    "$(need DEBUG)"
set_str "$S_TOP" data_dir "$CONTAINER_DATA_DIR"

set_raw "$S_GIT" create_deploy_key "$(need GIT_CREATE_DEPLOY_KEY)"
set_str "$S_GIT" repo             "$(need GIT_REPO)"
set_str "$S_GIT" branch           "$(need GIT_BRANCH)"
set_raw "$S_GIT" polling_interval "$(need_int GIT_POLLING_INTERVAL)"

set_str "$S_API" email            "$(need API_ADMIN_EMAIL)"
set_str "$S_API" password         "$(need API_ADMIN_PASSWORD)"
set_str "$S_API" secret           "$(need API_JWT_SECRET)"
set_raw "$S_API" expiration_hours "$(need_int API_JWT_EXPIRATION_HOURS)"
set_raw "$S_API" enable_api_spec  "$(need API_ENABLE_API_SPEC)"

set_str "$S_DB" host            "${DB_HOST:-db}"
set_raw "$S_DB" port            "$(need_int DB_PORT)"
set_str "$S_DB" user            "$(need DB_POSTGRES_USER)"
set_str "$S_DB" password        "$(need DB_POSTGRES_PASSWORD)"
set_str "$S_DB" dbname          "$(need DB_POSTGRES_DB)"
set_str "$S_DB" ssl_mode        "$(need DB_SSL_MODE)"
set_str "$S_DB" master_key      "$(need DB_MASTER_KEY)"
set_str "$S_DB" encryption_salt "$(need DB_ENCRYPTION_SALT)"

if grep -qE 'captainav0608@gmail\.com|<github_username>|"jwt-secret"|"qwertyuiop"' "$CONFIG"; then
    die "template placeholders remain in $CONFIG — a substitution did not apply"
fi

log "config rendered successfully"

exec "$@"

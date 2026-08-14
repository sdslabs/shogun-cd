#!/usr/bin/env sh
set -eu

# ──────────────────────────────────────────────
# Shogun-CD test cleanup
# Reverses everything setup.sh installs, so you
# can re-run the installer cleanly.
# ──────────────────────────────────────────────

BINARY="/usr/local/bin/shogun"
CONFIG_DIR="/etc/shogun"
DATA_DIR="/var/lib/shogun"
LOG_FILE="/var/log/shogun.log"
DB_CONTAINER="shogun_db"

echo "→ Stopping shogun process (if running)..."
PID=$(pgrep -f "$BINARY" 2>/dev/null || true)
if [ -n "$PID" ]; then
    kill "$PID" 2>/dev/null || true
    echo "  killed PID $PID"
else
    echo "  no running shogun process found"
fi

echo "→ Stopping Postgres container (if running)..."
if docker ps -a --format '{{.Names}}' 2>/dev/null | grep -q "^${DB_CONTAINER}\$"; then
    docker stop "$DB_CONTAINER" >/dev/null 2>&1 || true
    echo "  stopped and removed $DB_CONTAINER (--rm auto-cleans it)"
else
    echo "  no $DB_CONTAINER container found"
fi

echo "→ Removing installed binary..."
sudo rm -f "$BINARY"

echo "→ Removing config dir ($CONFIG_DIR)..."
sudo rm -rf "$CONFIG_DIR"

echo "→ Removing data dir ($DATA_DIR)..."
sudo rm -rf "$DATA_DIR"

echo "→ Removing log file ($LOG_FILE)..."
sudo rm -f "$LOG_FILE"

echo ""
echo "✓ Cleanup complete. System is back to pre-install state."
echo "  (Postgres image itself was left in place — remove with"
echo "   'docker rmi postgres:18.3-alpine3.23' if you want that gone too.)"
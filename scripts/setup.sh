#!/usr/bin/env sh
set -eu

# ──────────────────────────────────────────────
# Shogun-CD single-line setup
# curl -fsSL <raw_url> | sh
# ──────────────────────────────────────────────

RED='\033[31m'
GREEN='\033[32m'
YELLOW='\033[33m'
BLUE='\033[34m'
CYAN='\033[36m'
BOLD='\033[1m'
NC='\033[0m'

REPO="sdslabs/shogun-cd"
BINARY="/usr/local/bin/shogun"
CONFIG_DIR="/etc/shogun"
CONFIG_FILE="$CONFIG_DIR/config.yaml"
DATA_DIR="/var/lib/shogun"
DB_VOLUME="$DATA_DIR/.data/db_volume"
LOG_FILE="/var/log/shogun.log"

# ── helpers ──────────────────────────────────

print_status()  { printf "${BLUE}→${NC} %s\n" "$1"; }
print_success() { printf "${GREEN}✓${NC} %s\n" "$1"; }
print_warn()    { printf "${YELLOW}⚠${NC} %s\n" "$1"; }
print_error()   { printf "${RED}✗${NC} %s\n" "$1"; }
print_bold()    { printf "${BOLD}%s${NC}\n" "$1"; }

die() {
    print_error "$1"
    exit 1
}

need_sudo() {
    if [ "$(id -u)" -ne 0 ]; then
        printf "This step needs root privileges.\n"
        printf "Enter your password to continue: "
        sudo -v || die "sudo authentication failed"
    fi
}

random_hex() {
    openssl rand -hex "$1" 2>/dev/null || {
        # fallback if openssl is missing (shouldn't happen on macOS/Linux)
        hex=""
        i=0
        while [ "$i" -lt "$1" ]; do
            b=$(od -An -N1 -tu1 /dev/urandom 2>/dev/null | tr -d ' ')
            hex="${hex}$(printf '%02x' "$b")"
            i=$((i + 1))
        done
        echo "$hex"
    }
}

# ── banner ───────────────────────────────────

printf "\n"
printf "${CYAN}${BOLD}"
printf "  ╔══════════════════════════════════════╗\n"
printf "  ║         Shogun-CD Installer          ║\n"
printf "  ╚══════════════════════════════════════╝\n"
printf "${NC}\n"

# ── step 1: detect OS + ARCH ────────────────

print_status "Detecting system..."

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
    linux|darwin) ;;
    *) die "Unsupported OS: $OS (only Linux and macOS are supported)" ;;
esac

case "$ARCH" in
    x86_64|amd64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) die "Unsupported architecture: $ARCH" ;;
esac

print_success "Detected: $OS / $ARCH"

# ── step 2: fetch latest release tag ─────────

print_status "Fetching latest release..."

RELEASE_URL="https://api.github.com/repos/$REPO/releases/latest"
TAG=$(curl -fsSL "$RELEASE_URL" 2>/dev/null | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\(.*\)".*/\1/')

if [ -z "$TAG" ]; then
    print_warn "Could not fetch latest release tag from GitHub API"
    printf "Enter a release tag manually (e.g. v0.1.0): "
    read -r TAG
    [ -n "$TAG" ] || die "No release tag provided"
fi

print_success "Release: $TAG"

# ── step 3: download binary ─────────────────

BINARY_NAME="shogun_${OS}_${ARCH}"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$TAG/$BINARY_NAME"

print_status "Downloading shogun binary..."
print_status "URL: $DOWNLOAD_URL"

# Single tmp dir for everything downloaded/extracted below — cleaned up
# in one shot once the binary is installed (step 4).
TMP_DIR=$(mktemp -d)
TMP_BIN="$TMP_DIR/shogun"

curl -fsSL "$DOWNLOAD_URL" -o "$TMP_BIN" 2>/dev/null || {
    # Try alternate naming convention
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$TAG/shogun-cd_${OS}_${ARCH}"
    print_status "Trying alternate URL: $DOWNLOAD_URL"
    curl -fsSL "$DOWNLOAD_URL" -o "$TMP_BIN" || {
        # Try .tar.gz
        TAR_URL="https://github.com/$REPO/releases/download/$TAG/shogun-cd_${OS}_${ARCH}.tar.gz"
        print_status "Trying tar.gz: $TAR_URL"
        TMP_TAR="$TMP_DIR/shogun.tar.gz"
        curl -fsSL "$TAR_URL" -o "$TMP_TAR" || die "Failed to download binary. Check that release $TAG exists and has artifacts for $OS/$ARCH."
        tar -xzf "$TMP_TAR" -C "$TMP_DIR" shogun 2>/dev/null || tar -xzf "$TMP_TAR" -C "$TMP_DIR" 2>/dev/null || die "Failed to extract tar.gz"
        TMP_BIN=$(find "$TMP_DIR" -name "shogun" -type f 2>/dev/null | head -1)
        [ -n "$TMP_BIN" ] && [ -f "$TMP_BIN" ] || die "Could not find shogun binary in extracted archive"
    }
}

# ── step 4: install binary ──────────────────

print_status "Installing shogun to $BINARY..."

need_sudo
sudo mkdir -p "$(dirname "$BINARY")"
sudo mv "$TMP_BIN" "$BINARY"
sudo chmod +x "$BINARY"
rm -rf "$TMP_DIR"

print_success "Binary installed at $BINARY"

# ── step 5: create directories ──────────────

print_status "Creating directories..."

sudo mkdir -p "$CONFIG_DIR"
sudo mkdir -p "$DATA_DIR"
sudo mkdir -p "$DB_VOLUME"
# shogun runs unprivileged (see step 11) and writes its deploy key under
# $DATA_DIR/ssh/, so hand ownership to the invoking user.
sudo chown -R "$(id -u):$(id -g)" "$DATA_DIR"

print_success "Directories created"

# ── step 6: database setup ──────────────────

printf "\n${BOLD}Database Setup${NC}\n"
printf "Shogun-CD requires a PostgreSQL database.\n\n"

printf "Do you have an existing PostgreSQL database? [y/N]: "
read -r HAS_DB
HAS_DB=$(echo "$HAS_DB" | tr '[:upper:]' '[:lower:]')

if [ "$HAS_DB" = "y" ] || [ "$HAS_DB" = "yes" ]; then
    # ── existing database ──
    printf "\nProvide your database connection details:\n"

    printf "  Host [127.0.0.1]: "
    read -r DB_HOST
    DB_HOST=${DB_HOST:-127.0.0.1}

    printf "  Port [5432]: "
    read -r DB_PORT
    DB_PORT=${DB_PORT:-5432}

    printf "  User [shogun]: "
    read -r DB_USER
    DB_USER=${DB_USER:-shogun}

    printf "  Password: "
    stty -echo
    read -r DB_PASSWORD
    stty echo
    printf "\n"

    printf "  Database name [shogun_db]: "
    read -r DB_NAME
    DB_NAME=${DB_NAME:-shogun_db}

    DB_SSL="disable"

    print_success "Database configuration captured"
else
    # ── docker path ──
    if ! command -v docker >/dev/null 2>&1; then
        die "Docker is not installed. Install Docker or provide an existing PostgreSQL database and re-run the script."
    fi

    print_success "Docker found"

    printf "\nUse an auto-generated database password or provide your own?\n"
    printf "  [a] Auto-generate (recommended)\n"
    printf "  [m] Manual\n"
    printf "Choice [a]: "
    read -r PWD_CHOICE
    PWD_CHOICE=$(echo "$PWD_CHOICE" | tr '[:upper:]' '[:lower:]')

    DB_HOST="127.0.0.1"
    DB_PORT="5432"
    DB_USER="shogun"
    DB_NAME="shogun_db"
    DB_SSL="disable"

    if [ "$PWD_CHOICE" = "m" ]; then
        printf "  Database password: "
        stty -echo
        read -r DB_PASSWORD
        stty echo
        printf "\n"
    else
        DB_PASSWORD=$(random_hex 16)
        printf "\n"
        print_status "Auto-generated DB password: ${BOLD}$DB_PASSWORD${NC}"
    fi

    print_status "Starting PostgreSQL container..."

    # Stop existing container if present
    docker stop shogun_db 2>/dev/null || true

    docker pull postgres:18.3-alpine3.23 >/dev/null 2>&1

    docker run --name shogun_db --rm \
        -v "$DB_VOLUME:/var/lib/postgresql/" \
        -p "$DB_PORT:5432" \
        -e "POSTGRES_USER=$DB_USER" \
        -e "POSTGRES_PASSWORD=$DB_PASSWORD" \
        -e "POSTGRES_DB=$DB_NAME" \
        -d postgres:18.3-alpine3.23 >/dev/null 2>&1

    print_success "PostgreSQL container started (port $DB_PORT)"

    # Wait for DB to be ready
    print_status "Waiting for database to be ready..."
    for _ in $(seq 1 30); do
        if docker exec shogun_db pg_isready -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1; then
            break
        fi
        sleep 1
    done
    print_success "Database is ready"
fi

# ── step 7: git repo ────────────────────────

printf "\n${BOLD}Git Repository${NC}\n"
printf "Shogun-CD watches a Git repository for Pipeline and Target manifests.\n\n"

printf "  Git repo URL (SSH recommended, e.g. git@github.com:user/repo.git): "
read -r GIT_REPO
[ -n "$GIT_REPO" ] || die "Git repository URL is required"

printf "  Branch [main]: "
read -r GIT_BRANCH
GIT_BRANCH=${GIT_BRANCH:-main}

# ── step 8: admin account ───────────────────

printf "\n${BOLD}Admin Account${NC}\n"
printf "This account will have full access to the Shogun-CD API.\n\n"

printf "  Admin email: "
read -r ADMIN_EMAIL
[ -n "$ADMIN_EMAIL" ] || die "Admin email is required"

printf "  Admin password: "
stty -echo
read -r ADMIN_PASSWORD
stty echo
printf "\n"
[ -n "$ADMIN_PASSWORD" ] || die "Admin password is required"

# ── step 9: generate secrets ────────────────

MASTER_KEY=$(random_hex 16)
ENCRYPTION_SALT=$(random_hex 8)
JWT_SECRET=$(random_hex 16)

# ── step 10: write config ───────────────────

print_status "Writing configuration to $CONFIG_FILE..."

sudo tee "$CONFIG_FILE" > /dev/null <<YAML_EOF
debug: true

data_dir: "$DATA_DIR/"

git:
    create_deploy_key: true
    repo: "$GIT_REPO"
    branch: "$GIT_BRANCH"
    polling_interval: 5

api:
    admin:
        email: "$ADMIN_EMAIL"
        password: "$ADMIN_PASSWORD"
    port: 7007
    jwt:
        secret: "$JWT_SECRET"
        expiration_hours: 1
    enable_api_spec: true

database:
    host: "$DB_HOST"
    port: $DB_PORT
    user: "$DB_USER"
    password: "$DB_PASSWORD"
    dbname: "$DB_NAME"
    ssl_mode: "$DB_SSL"
    master_key: "$MASTER_KEY"
    encryption_salt: "$ENCRYPTION_SALT"
YAML_EOF

print_success "Configuration written"

# ── step 11: start shogun ───────────────────

print_status "Starting Shogun-CD..."

# Kill any existing shogun process
pkill -f "$BINARY" 2>/dev/null || true
sleep 1

# /var/log is root-owned — create the log file as root, then hand it to
# the invoking user so shogun (running unprivileged, below) can write to
# it directly without needing sudo at runtime.
sudo touch "$LOG_FILE"
sudo chown "$(id -u):$(id -g)" "$LOG_FILE"

nohup "$BINARY" > "$LOG_FILE" 2>&1 &
SHOGUN_PID=$!

sleep 2  # give it a moment to start (or fail)

if ! kill -0 "$SHOGUN_PID" 2>/dev/null; then
    print_error "Shogun-CD failed to start. Last lines from $LOG_FILE:"
    tail -n 20 "$LOG_FILE" 2>/dev/null || true
    die "Startup failed — see log above."
fi

# ── step 12: summary ────────────────────────

sleep 2  # Give shogun a moment to generate deploy keys

printf "\n"
printf "${GREEN}${BOLD}"
printf "  ╔══════════════════════════════════════╗\n"
printf "  ║     Shogun-CD is now running!        ║\n"
printf "  ╚══════════════════════════════════════╝\n"
printf "${NC}\n"

printf "  ${BOLD}API:${NC}      http://localhost:7007\n"
printf "  ${BOLD}Docs:${NC}     http://localhost:7007/docs/\n"
printf "  ${BOLD}Config:${NC}   $CONFIG_FILE\n"
printf "  ${BOLD}Logs:${NC}    $LOG_FILE\n"
printf "  ${BOLD}Data:${NC}    $DATA_DIR/\n"
printf "\n"
printf "  ${BOLD}Database credentials:${NC}\n"
printf "    Host:     $DB_HOST\n"
printf "    Port:     $DB_PORT\n"
printf "    User:     $DB_USER\n"
printf "    Password: $DB_PASSWORD\n"
printf "    Database: $DB_NAME\n"

# Deploy key guidance
printf "\n"
printf "  ${BOLD}Deploy Key:${NC}\n"
printf "  Shogun generates an ed25519 key pair for Git repo access.\n"
printf "  Expected at: ${YELLOW}$DATA_DIR/ssh/deploy_id_ed25519.pub${NC}\n"
printf "  Add that public key as a deploy key to your Git repo.\n"

printf "\n"
printf "  ${BOLD}Process ID:${NC} $SHOGUN_PID\n"
printf "  ${BOLD}Logs:${NC}    $LOG_FILE\n"
printf "  Stop with:  ${CYAN}kill $SHOGUN_PID${NC}\n"
printf "\n"
print_success "Setup complete!"
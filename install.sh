#!/usr/bin/env bash

set -e

REPO="sdslabs/shogun-cd"
BRANCH="develop" # Default branch to use for URLs
CONFIG_DIR="/etc/shogun"
DATA_DIR="/var/lib/shogun"
DB_VOLUME="${DATA_DIR}/db_volume"
BIN_DIR="/usr/local/bin"
SERVICE_FILE="/etc/systemd/system/shogun.service"
START_SHOGUN_POSTGRES=${START_SHOGUN_POSTGRES:-true}

# Postgres configuration
DB_CONTAINER_NAME="shogun_db"
DB_USER="shogun"
DB_PASSWORD="pwd"
DB_NAME="shogun_db"
DB_PORT_MAPPING="5432:5432"

# Color formatting
CYAN='\033[0;36m'
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() { echo -e "${CYAN}==>${NC} $1"; }
success() { echo -e "${GREEN}==>${NC} $1"; }
warn() { echo -e "${YELLOW}==>${NC} $1"; }
error() { echo -e "${RED}==>${NC} $1"; exit 1; }

# Validations
if [ "$(id -u)" != "0" ]; then
    error "This script must be run as root. Try running: curl -sSL <url> | sudo bash"
fi

if [ "$(uname -s)" != "Linux" ]; then
    error "This script only supports Linux."
fi

ARCH=$(uname -m)
case $ARCH in
    x86_64) OS_ARCH="amd64" ;;
    aarch64|arm64) OS_ARCH="arm64" ;;
    *) error "Unsupported architecture: $ARCH" ;;
esac

log "Detected architecture: $OS_ARCH"

# Dependencies
install_package() {
    local pkg=$1
    log "$pkg is not installed. Attempting to install..."
    if command -v apt-get >/dev/null 2>&1; then
        apt-get install -y "$pkg" || { warn "Failed to install $pkg. You may need to run 'apt-get update' manually."; error "Failed to install $pkg."; }
    elif command -v yum >/dev/null 2>&1; then
        yum install -y "$pkg"
    elif command -v apk >/dev/null 2>&1; then
        apk add "$pkg"
    else
        error "Could not install $pkg automatically. Please install $pkg and try again."
    fi
}

command -v curl >/dev/null 2>&1 || install_package curl
command -v tar >/dev/null 2>&1 || install_package tar
command -v git >/dev/null 2>&1 || install_package git

# Database Setup
if [ "$START_SHOGUN_POSTGRES" = "true" ]; then
    log "Setting up PostgreSQL using Docker..."
    if ! command -v docker >/dev/null 2>&1; then
        error "Docker is not installed. Required to start the database. Please install Docker, or set START_SHOGUN_POSTGRES=false to skip."
    fi
    
    mkdir -p "${DB_VOLUME}"
    
    if grep -q "${DB_CONTAINER_NAME}" <<< "$(docker ps -a --format '{{.Names}}')"; then
        warn "Docker container '${DB_CONTAINER_NAME}' already exists. Skipping database startup."
    else
        docker run --name ${DB_CONTAINER_NAME} \
            --rm \
            -v "${DB_VOLUME}:/var/lib/postgresql/data" \
            -p ${DB_PORT_MAPPING} \
            -e POSTGRES_USER=${DB_USER} \
            -e POSTGRES_PASSWORD=${DB_PASSWORD} \
            -e POSTGRES_DB=${DB_NAME} \
            -d postgres:18.3-alpine3.23
        success "Database started successfully!"
    fi
else
    log "Skipping Docker Postgres startup."
fi

# Download Binary
log "Fetching latest Shogun release..."
# Using github api to get latest release
LATEST_URL=$(curl -s https://api.github.com/repos/${REPO}/releases/latest | grep "browser_download_url" | grep "linux_${OS_ARCH}" | cut -d '"' -f 4 || true)

# Fallback: if there are no releases yet, we might need a placeholder or instruct to build from source
if [ -z "$LATEST_URL" ]; then
    warn "No releases found for ${REPO}. Please build from source or check back later."
    exit 0
fi

TMP_BIN="/tmp/shogun.tar.gz"
curl -sSL "$LATEST_URL" -o "$TMP_BIN"
tar -xzf "$TMP_BIN" -C /tmp
mv /tmp/shogun ${BIN_DIR}/shogun
rm -f "$TMP_BIN"
chmod +x ${BIN_DIR}/shogun
success "Shogun installed at ${BIN_DIR}/shogun"

# Setup Directories and Config
log "Setting up configuration in ${CONFIG_DIR}..."
mkdir -p "${CONFIG_DIR}"
mkdir -p "${DATA_DIR}"

if [ ! -f "${CONFIG_DIR}/config.yaml" ]; then
    # Download sample config
    SAMPLE_URL="https://raw.githubusercontent.com/${REPO}/${BRANCH}/sample.config.yaml"
    curl -sSL "$SAMPLE_URL" -o "${CONFIG_DIR}/config.yaml" || true
    
    # In case curl failed (e.g. branch is not main), try to write a default if empty
    if [ ! -s "${CONFIG_DIR}/config.yaml" ]; then
        echo 'git:
    repo: "git@github.com:sdslabs/shogun-manifest-demo.git"
    branch: "main"' > "${CONFIG_DIR}/config.yaml"
    fi

    # Read from /dev/tty because curl | bash consumes stdin
    printf "${CYAN}==>${NC} Enter the GitHub repository URL to track (leave blank to skip, you can modify it later and restart shogun service): "
    read -r user_repo < /dev/tty || true
    
    if [ -n "$user_repo" ]; then
        sed -i "s|git@github.com:sdslabs/shogun-manifest-demo.git|$user_repo|g" "${CONFIG_DIR}/config.yaml"
        success "Repository set to: $user_repo"
    else
        warn "No repository provided. You can update ${CONFIG_DIR}/config.yaml later."
    fi
else
    warn "Configuration file ${CONFIG_DIR}/config.yaml already exists. Skipping overwrite."
fi

# Systemd Service
log "Setting up systemd service..."
cat <<EOF > ${SERVICE_FILE}
[Unit]
Description=Shogun CD Tool
After=network.target

[Service]
Type=simple
ExecStart=${BIN_DIR}/shogun
Restart=always
RestartSec=5
WorkingDirectory=${DATA_DIR}

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable shogun
systemctl start shogun

success "Shogun-CD installed and started!"
log "View logs with: journalctl -u shogun -f"

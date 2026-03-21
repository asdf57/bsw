#!/usr/bin/env bash
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "Run this script as root (sudo)." >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

if [[ -z "${DOMAIN:-}" ]]; then
  echo "Missing DOMAIN env var (e.g. DOMAIN=api.example.com)." >&2
  exit 1
fi

if [[ -z "${CF_API_TOKEN:-}" ]]; then
  echo "Missing CF_API_TOKEN env var." >&2
  exit 1
fi

echo "[1/5] Installing Docker..."
apt-get update -qq
apt-get install -y -qq ca-certificates curl gnupg git ufw
install -m 0755 -d /etc/apt/keyrings
if [[ ! -f /etc/apt/keyrings/docker.gpg ]]; then
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
fi
chmod a+r /etc/apt/keyrings/docker.gpg
if [[ ! -f /etc/apt/sources.list.d/docker.list ]]; then
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "${VERSION_CODENAME}") stable" > /etc/apt/sources.list.d/docker.list
fi
apt-get update -qq
apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-compose-plugin

if [[ ! -f "${APP_DIR}/docker-compose.yml" || ! -f "${APP_DIR}/cloud/docker-compose.yml" ]]; then
  echo "Could not find repo files from ${APP_DIR}." >&2
  echo "Run this script from inside the bsw repository checkout on the droplet." >&2
  exit 1
fi

echo "[2/5] Writing cloud env file..."
cat > "${APP_DIR}/cloud/.env" <<EOF
DOMAIN=${DOMAIN}
CF_API_TOKEN=${CF_API_TOKEN}
POSTGRES_USER=${POSTGRES_USER:-user}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-password}
POSTGRES_DB=${POSTGRES_DB:-bswdb}
POSTGRES_PORT=${POSTGRES_PORT:-5433}
EOF

echo "[3/5] Opening firewall ports (SSH, HTTP, HTTPS)..."
ufw --force reset
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

echo "[4/5] Deploying cloud stack..."
docker compose --env-file "${APP_DIR}/cloud/.env" -f "${APP_DIR}/cloud/docker-compose.yml" pull || true
docker compose --env-file "${APP_DIR}/cloud/.env" -f "${APP_DIR}/cloud/docker-compose.yml" up -d --build

echo "[5/5] Deployment complete."
echo "Check status with:"
echo "  docker compose --env-file ${APP_DIR}/cloud/.env -f ${APP_DIR}/cloud/docker-compose.yml ps"

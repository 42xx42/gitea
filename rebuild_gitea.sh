#!/bin/bash
# rebuild_gitea.sh — Build custom Gitea with audit+watermark on the server
# Run as root on server 5.175.233.43
set -euo pipefail

INSTALL_GO=false
BUILD_GITEA=true
DEPLOY=true

# ──────────────────────────────────────────────
# 1. Install Go 1.24+ if needed
# ──────────────────────────────────────────────
if ! command -v go &>/dev/null || [[ $(go version | grep -oP 'go\K[0-9]+\.[0-9]+' | head -1) < "1.24" ]]; then
    echo ">>> Installing Go 1.24.3 ..."
    wget -q https://go.dev/dl/go1.24.3.linux-amd64.tar.gz -O /tmp/go1.24.3.linux-amd64.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf /tmp/go1.24.3.linux-amd64.tar.gz
    export PATH=/usr/local/go/bin:$PATH
    echo 'export PATH=/usr/local/go/bin:$PATH' >> /etc/profile.d/go.sh
    rm /tmp/go1.24.3.linux-amd64.tar.gz
fi
export PATH=/usr/local/go/bin:$PATH
echo "Go version: $(go version)"

# ──────────────────────────────────────────────
# 2. Get our custom source
# ──────────────────────────────────────────────
GITEA_SRC="/opt/gitea-audit-src"
if [ ! -d "$GITEA_SRC" ]; then
    echo ">>> Cloning custom Gitea fork ..."
    git clone https://github.com/42xx42/gitea.git "$GITEA_SRC"
fi
cd "$GITEA_SRC"
git fetch origin
git checkout master
git reset --hard origin/master

# ──────────────────────────────────────────────
# 3. Build with bindata tags (critical!)
# ──────────────────────────────────────────────
echo ">>> Building Gitea with bindata + sqlite tags ..."
export CGO_ENABLED=1
export TAGS="bindata sqlite sqlite_unlock_notify"
make build

# The binary will be at gitea
ls -lh gitea
echo "Build complete!"

# ──────────────────────────────────────────────
# 4. Copy into Docker image
# ──────────────────────────────────────────────
echo ">>> Creating Docker image with new binary ..."

# Stop the current container
cd /srv/gitea-stack
docker compose stop gitea 2>/dev/null || true

# Create a temp container from original image, replace binary, commit
docker create --name gitea-rebuild docker.gitea.com/gitea:1.24.7-rootless
docker cp "$GITEA_SRC/gitea" gitea-rebuild:/app/gitea/gitea
docker commit gitea-rebuild gitea-audit:1.24.7-audit
docker rm gitea-rebuild

echo ">>> New image created: gitea-audit:1.24.7-audit"

# ──────────────────────────────────────────────
# 5. Fix docker-compose.yml to use our image
# ──────────────────────────────────────────────
COMPOSE_FILE="/srv/gitea-stack/docker-compose.yml"
if grep -q 'docker.gitea.com/gitea:1.24.7-rootless' "$COMPOSE_FILE"; then
    sed -i 's|docker.gitea.com/gitea:1.24.7-rootless|gitea-audit:1.24.7-audit|g' "$COMPOSE_FILE"
    echo ">>> Updated docker-compose.yml to use custom image"
fi

# ──────────────────────────────────────────────
# 6. Start
# ──────────────────────────────────────────────
echo ">>> Starting Gitea ..."
docker compose up -d gitea

# Wait and check
sleep 5
if docker ps | grep -q gitea; then
    echo "✅ Gitea is running!"
    docker logs --tail 20 git42w-gitea 2>&1 | tail -20
else
    echo "❌ Gitea failed to start. Checking logs:"
    docker logs --tail 50 git42w-gitea 2>&1 | tail -50
fi

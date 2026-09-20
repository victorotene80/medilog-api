#!/usr/bin/env bash
# Run ON the Oracle VM (ssh in first) from the repo root to ship a new
# version: pulls latest main, rebuilds the image in place, applies
# migrations, then restarts the stack. No registry, no CI secrets — the VM
# builds its own image, which is also why it's a native arm64 build with no
# cross-compilation involved.
#
# Usage: ./scripts/deploy.sh
set -euo pipefail

COMPOSE="docker compose -f docker-compose.prod.yml"

if [ ! -f .env ]; then
  echo "ERROR: .env not found. Copy .env.prod.example to .env and fill it in first." >&2
  exit 1
fi

echo "=== Pulling latest code ==="
git fetch origin
git reset --hard origin/main

echo "=== Building image ==="
$COMPOSE build api

echo "=== Applying migrations ==="
$COMPOSE run --rm migrate

echo "=== Restarting stack ==="
$COMPOSE up -d postgres redis api caddy
if grep -q '^SCHEDULER_ENABLED=true' .env; then
  $COMPOSE --profile scheduler up -d scheduler-cron
fi

echo "=== Pruning old images ==="
docker image prune -f

echo "=== Recent API logs ==="
$COMPOSE logs --tail=30 api

echo "=== Done ==="

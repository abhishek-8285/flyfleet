#!/bin/bash
# docker-clean.sh - Maintenance script to keep storage healthy

echo "🧹 Starting Docker cleanup..."

# Remove dangling images, builders, and unused containers
docker system prune -f

# Remove all unused images older than 7 days
docker image prune -a -f --filter "until=168h"

# Optional: Clean builder cache
docker builder prune -f

echo "✨ Cleanup complete!"
df -h /

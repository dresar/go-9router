#!/usr/bin/env bash
set -e
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$DIR"

echo "Starting Go 9Router on port 20128..."
exec ./bin/9router -p 20128 -n

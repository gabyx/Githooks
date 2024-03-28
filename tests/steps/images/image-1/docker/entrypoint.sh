#!/usr/bin/env bash
set -u
set -e

echo "Containerized hook: entrypoint ==================="
echo "Working Dir: $(pwd)"
echo "User: $(id)"

echo "Permissions for '.':"
ls -ald "$(pwd)"
ls -al "$(pwd)"

echo "Permissions for /mnt/workspace and /mnt/shared"
ls -ald /mnt/workspace
ls -ald /mnt/shared || true
echo "=================================================="

echo "Launching inside container:" "$@"
exec "$@"

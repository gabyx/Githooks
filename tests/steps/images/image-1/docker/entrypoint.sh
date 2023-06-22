#!/usr/bin/env bash
set -u
set -e

echo "Containerized hook: entrypoint ==================="
echo "Working Dir: $(pwd)"
echo "User: $(id)"
echo "Permissions for '.':"
ls -al .
echo "=================================================="

echo "Lanuching" "$@"

exec "$@"

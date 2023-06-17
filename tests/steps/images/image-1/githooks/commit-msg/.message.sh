#!/bin/bash
set -u
set -e

echo "Arguments given:" "$@"

if [ "${1:-}" != "--message" ]; then
    echo "! First argument is not --message"
    exit 1
fi

echo "Containerized commit-msg hook run" > ./.commit-msg-hook-run

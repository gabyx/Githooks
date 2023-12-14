#!/bin/bash
set -u
set -e

echo "Arguments given:" "$@"

if [ -z "${1:-}" ] || [ "${1:-}" == "--message" ]; then
    echo "! First argument must be the file to 'commit-msg' hook."
    exit 1
fi

if [ "${2:-}" != "--message" ]; then
    echo "! Second argument is not --message"
    exit 1
fi

echo "Containerized commit-msg hook run" >./.commit-msg-hook-run

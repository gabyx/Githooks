#!/bin/bash
set -u
set -e

echo "Arguments given:" "$@"

if [ "${1:-}" != "--message" ]; then
    echo "! First argument is not --message"
    exit 1
fi

if [ -z "${2:-}" ]; then
    echo "! Second argument must be the file to 'commit-msg' hook."
    exit 1
fi

echo "Containerized commit-msg hook run" >./.commit-msg-hook-run

#!/bin/sh
# Test:
#   Cli tool: add/update README

if ! sh /var/lib/githooks/install.sh; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p /tmp/test080 && cd /tmp/test080 && git init || exit 1

sh /var/lib/githooks/cli.sh readme update &&
    [ -f .githooks/README.md ] ||
    exit 1

if sh /var/lib/githooks/cli.sh readme add; then
    echo "! Expected to fail"
    exit 1
fi

# Check the Git alias
rm -f .githooks/README.md &&
    git hooks readme add &&
    [ -f .githooks/README.md ] ||
    exit 1
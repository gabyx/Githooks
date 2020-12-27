#!/bin/sh
# Test:
#   Run the cli tool trying to list hooks of invalid type

if ! "$GITHOOKS_BIN_DIR/installer" --stdin; then
    echo "! Failed to execute the install script"
    exit 1
fi

mkdir -p /tmp/test072/.githooks/pre-commit &&
    echo 'echo "Hello"' >/tmp/test072/.githooks/pre-commit/testing &&
    cd /tmp/test072 &&
    git init ||
    exit 1

if ! git hooks list pre-commit; then
    echo "! Failed to execute a valid list"
    exit 1
fi

if ! git hooks list invalid-type | grep -i 'no active hooks'; then
    echo "! Unexpected list result"
    exit 1
fi

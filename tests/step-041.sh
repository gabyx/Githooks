#!/bin/sh
# Test:
#   Run a single-repo, dry-run install successfully

if echo "$EXTRA_INSTALL_ARGS" | grep -q "use-core-hookspath"; then
    echo "Using core.hooksPath"
    exit 249
fi

mkdir -p /tmp/start/dir && cd /tmp/start/dir || exit 1

git init || exit 1

if ! "$GITHOOKS_TEST_BIN_DIR/installer" --dry-run; then
    echo "! Installation failed"
    exit 1
fi

if grep -r 'github.com/rycus86/githooks' /tmp/start/dir/.git/hooks; then
    echo "! Hooks were not expected to be installed"
    exit 1
fi

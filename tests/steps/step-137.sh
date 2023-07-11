#!/usr/bin/env bash
# Test:
#   Run install.sh script from the main branch.
set -u

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091

. "$TEST_DIR/general.sh"

# Install with script.
curl -sL https://raw.githubusercontent.com/gabyx/githooks/main/scripts/install.sh | bash -s -- -- \
    --use-core-hookspath

mkdir -p "$GH_TEST_TMP/test136" &&
    cd "$GH_TEST_TMP/test136" &&
    git init

if [ -z "$(git config core.hooksPath)" ]; then
    echo "Git core.hooskPath is not set but should."
    exit 1
fi

if [ -n "$(git config init.templateDir)" ]; then
    echo "Git init.templateDir is set but should not."
    exit 1
fi

if ! grep -Rq 'github.com/gabyx/githooks' .git/hooks; then
    echo "Hooks should not have been installed."
    exit 1
fi

git hooks uninstaller || exit 1
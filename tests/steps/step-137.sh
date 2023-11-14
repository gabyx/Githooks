#!/usr/bin/env bash
# Test:
#   Run install.sh script from the main branch.
set -u

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091

. "$TEST_DIR/general.sh"

if [ -n "${GH_COVERAGE_DIR:-}" ]; then
    echo "Test cannot run for coverage."
    exit 249
fi

ref="${CIRCLE_SHA1:-main}"
# Install with script.
curl -sL "https://raw.githubusercontent.com/gabyx/Githooks/$ref/scripts/install.sh" | bash -s -- -- \
    --use-core-hookspath

mkdir -p "$GH_TEST_TMP/test137" &&
    cd "$GH_TEST_TMP/test137" &&
    git init

if [ -z "$(git config core.hooksPath)" ]; then
    echo "Git core.hooskPath is not set but should."
    exit 1
fi

if [ -n "$(git config init.templateDir)" ]; then
    echo "Git init.templateDir is set but should not."
    exit 1
fi

if grep -Rq 'github.com/gabyx/githooks' .git/hooks; then
    echo "Hooks should not have been installed."
    exit 1
fi

git hooks uninstaller || exit 1

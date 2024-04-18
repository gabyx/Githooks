#!/usr/bin/env bash
# Test:
#   Run install.sh script.
set -u

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091

. "$TEST_DIR/general.sh"

init_step

if [ -n "${GH_COVERAGE_DIR:-}" ]; then
    echo "Test cannot run for coverage."
    exit 249
fi

# Clone repo on feature/v3 to add a fictuous version 3.0.0 to test update
# failure.
# You can change this to `main` once `feature/v3` is deleted.
BRANCH="feature/v3"

mkdir -p "$GH_TEST_TMP/test-146" &&
    cd "$GH_TEST_TMP/test-146" &&
    git clone --single-branch --branch "$BRANCH" https://github.com/gabyx/githooks.git ||
    exit 1

# Install with current script the version 2.10.0 on the `feature/v3` branch.
curl -sL "file://$GH_SCRIPTS/install.sh" | bash -s -- --version 2.10.0 -- --clone-branch "feature/v3" || {
    echo "Could not download install.sh and install 2.10.0 on branch '$BRANCH'."
    exit 1
}
# Uninstall right away again.
"$GH_INSTALL_BIN_DIR/cli" uninstaller || exit 1

#!/usr/bin/env bash
# Test:
#   Download from deploy url and install latest version
# shellcheck disable=SC1091

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

git -C "$GH_TEST_REPO" reset --hard v3.0.0 >/dev/null 2>&1 || exit 1

# We do not want to inject special flags into the downloaded executable because it has not been build with coverage
# therefore -> use `GH_DEPLOY_SOURCE_IS_PROD` in `executables-coverage.go`
GH_DEPLOY_SOURCE_IS_PROD=true \
    "$GH_TEST_BIN/githooks-cli" installer \
    --clone-url "https://github.com/gabyx/Githooks.git" \
    --clone-branch "main" \
    exit 1

"$GH_INSTALL_BIN_DIR/githooks-cli" uninstaller

# Remove this installation, such that the uninstall in exec-steps works.
rm -rf "$GH_INSTALL_DIR" || exit 1

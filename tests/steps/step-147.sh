#!/usr/bin/env bash
# Test:
#   Test package-manager enabled build with run-wrappers not using githooks.runner

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

if [ -n "$GH_ON_WINDOWS" ]; then
    echo "On windows somehow this test gets stuck (?)."
    exit 249
fi

init_step

# Use the 3.1.0 prod build with package-manager enabled and download mocked.
git -C "$GH_TEST_REPO" reset --hard v3.1.0 >/dev/null 2>&1 || exit 1

# run the default install
"$GH_TEST_BIN/githooks-cli" installer \
    "${EXTRA_INSTALL_ARGS[@]}" \
    --clone-url "file://$GH_TEST_REPO" \
    --clone-branch "test-package-manager" || exit 1

# Test run-wrappers with pure binaries in the path.
# Put binaries into the path to find them.
export PATH="$GH_TEST_BIN:$PATH"
if [ -z "$GH_COVERAGE_DIR" ]; then
    # Coverage build will not run this because its wrapped...
    githooks-cli --version || {
        echo "! Binaries not in path."
        exit 1
    }
fi

# Overwrite runner, it should not be needed.
git config --global --unset githooks.runner

# We need this CLI at the right place for further tests.
[ ! -f ~/.githooks/bin/githooks-cli ] || {
    echo "! githooks-cli should not exist in default location"
    exit 1
}
# Install it into the default location for the test functions...
mkdir ~/.githooks/bin &&
    cp "$(which githooks-cli)" ~/.githooks/bin/ || exit 1

# Test the the runner works.
mkdir -p "$GH_TEST_TMP/test147" &&
    cd "$GH_TEST_TMP/test147" &&
    mkdir -p .githooks/pre-commit &&
    echo "echo 'Hook-1' >> '$GH_TEST_TMP/test'" >.githooks/pre-commit/test1 &&
    git init &&
    install_hooks_if_not_centralized || exit 1

if ! is_centralized_tests; then
    check_local_install
else
    check_centralized_install
fi

git hooks trust
git commit --allow-empty -m "Test hook" || exit 1

if [ ! -f "$GH_TEST_TMP/test" ]; then
    echo "! Hook did not run."
    exit 1
fi

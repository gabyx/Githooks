#!/usr/bin/env bash
# Test:
#   Test registering mechanism (centralized).
# shellcheck disable=SC1091

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

if ! echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    echo "Not using centralized install"
    exit 249
fi

accept_all_trust_prompts || exit 1

if ! "$GH_TEST_BIN/githooks-cli" installer; then
    echo "! Failed to execute the install script"
    exit 1
fi

REGISTER_FILE=~/.githooks/registered.yaml

if grep -q "/" "$REGISTER_FILE"; then
    echo "Expected the file to not contain any paths"
    exit 1
fi

# Test that first git action registers repo 1
mkdir -p "$GH_TEST_TMP/test144.1" && cd "$GH_TEST_TMP/test144.1" &&
    git init &&
    git commit --allow-empty -m 'Initial commit' &>/dev/null ||
    exit 1

if [ ! -f "$REGISTER_FILE" ]; then
    echo "Expected the file to be created"
    exit 1
fi

if ! grep -qE '.+/test *144.1/.git$' "$REGISTER_FILE"; then
    echo "Expected correct content"
    cat "$REGISTER_FILE"
    exit 2
fi

# Test that a first git action registers repo 2
# and repo 1 is still registered
mkdir -p "$GH_TEST_TMP/test144.2" && cd "$GH_TEST_TMP/test144.2" &&
    git init &&
    git commit --allow-empty -m 'Initial commit' &>/dev/null ||
    exit 1

if ! grep -qE '.+/test *144.1/.git$' "$REGISTER_FILE" ||
    ! grep -qE '.+/test *144.2/.git$' "$REGISTER_FILE"; then
    echo "! Expected correct content"
    cat "$REGISTER_FILE"
    exit 3
fi

mkdir -p "$GH_TEST_TMP/test144.3" &&
    cd "$GH_TEST_TMP/test144.3" &&
    git init || exit 1

# Should not have registered in repo 3
if ! grep -qE '.+/test *144.1/.git$' "$REGISTER_FILE" ||
    ! grep -qE '.+/test *144.2/.git$' "$REGISTER_FILE"; then
    echo "! Expected correct content"
    cat "$REGISTER_FILE"
    exit 3
fi

# Test uninstall to only repo 1
cd "$GH_TEST_TMP/test144.1" || exit 1
if ! "$GH_TEST_BIN/githooks-cli" uninstall; then
    echo "! Uninstall from current repo failed"
    exit 1
fi

if [ "$(git config --local githooks.registered)" = "true" ]; then
    echo "! Expected repo 1 to be marked unregistered"
    exit 1
fi

if grep -qE ".+/test *144.1/.git$" "$REGISTER_FILE" ||
    (! grep -qE ".+/test *144.2/.git$" "$REGISTER_FILE" &&
        ! grep -qE ".+/test *144.3/.git$" "$REGISTER_FILE"); then
    echo "! Expected repo 2 and 3 to still be registered"
    cat "$REGISTER_FILE"
    exit 5
fi

# Test total uninstall to all repos
echo "- Total uninstall..."
cd "$GH_TEST_TMP/test144.1" && git config githooks.runnerIsNonInteractive true
echo "Y
$GH_TEST_TMP
" | "$GH_TEST_BIN/githooks-cli" uninstaller --full-uninstall-from-repos --stdin || exit 1

if [ -f "$REGISTER_FILE" ]; then
    echo "! Expected registered list to not exist"
    exit 1
fi

if [ -n "$(git -C "$GH_TEST_TMP/test144.1" config githooks.runnerIsNonInteractive)" ]; then
    echo "! Expected to have cleaned the full Git config."
    exit 1
fi

check_no_install

# Reinstall everywhere
echo "- Reinstall everywhere..."
echo "y

Y
y
$GH_TEST_TMP
" | "$GH_TEST_BIN/githooks-cli" installer --stdin || exit 1

# Trigger the update only from repo 3
CURRENT_TIME=$(date +%s)
MOCK_LAST_RUN=$((CURRENT_TIME - 100000))

# Reset to trigger update from repo 3
if ! git -C "$GH_TEST_REPO" reset --hard v9.9.1 >/dev/null; then
    echo "! Could not reset master to trigger update."
    exit 1
fi

cd "$GH_TEST_TMP/test144.3" &&
    git config --global githooks.updateCheckEnabled true &&
    set_update_check_timestamp $MOCK_LAST_RUN &&
    OUT=$(git commit --allow-empty -m 'Second commit' 2>&1) || exit 1

if ! echo "$OUT" | grep -q "There is a new Githooks update available"; then
    echo "! Expected update-check output not found"
    echo "$OUT"
    exit 1
fi

OUT=$(git hooks update 2>&1)

if ! echo "$OUT" | grep -q "All done! Enjoy!"; then
    echo "! Expected installation output not found"
    echo "$OUT"
    exit 1
fi

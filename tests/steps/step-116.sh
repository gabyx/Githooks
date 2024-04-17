#!/usr/bin/env bash
# Test:
#   Test registering mechanism.
# shellcheck disable=SC1091

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    echo "Using centralized install"
    exit 249
fi

accept_all_trust_prompts || exit 1

if ! "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}"; then
    echo "! Failed to execute the install script"
    exit 1
fi

REGISTER_FILE=~/.githooks/registered.yaml

if grep -q "/" "$REGISTER_FILE"; then
    echo "Expected the file to not contain any paths"
    exit 1
fi

# Test that first git action registers repo 1
mkdir -p "$GH_TEST_TMP/test116.1" && cd "$GH_TEST_TMP/test116.1" &&
    git init &&
    "$GH_INSTALL_BIN_DIR/githooks-cli" install --maintained-hooks "all" &&
    git commit --allow-empty -m 'Initial commit' &>/dev/null ||
    exit 1

if [ ! -f "$REGISTER_FILE" ]; then
    echo "Expected the file to be created"
    exit 1
fi

if ! grep -qE '.+/test *116.1/.git$' "$REGISTER_FILE"; then
    echo "Expected correct content"
    cat "$REGISTER_FILE"
    exit 2
fi

# Test that a first git action registers repo 2
# and repo 1 is still registered
mkdir -p "$GH_TEST_TMP/test116.2" && cd "$GH_TEST_TMP/test116.2" &&
    git init &&
    "$GH_INSTALL_BIN_DIR/githooks-cli" install --maintained-hooks "all" &&
    git commit --allow-empty -m 'Initial commit' &>/dev/null ||
    exit 1

if ! grep -qE '.+/test *116.1/.git$' "$REGISTER_FILE" ||
    ! grep -qE '.+/test *116.2/.git$' "$REGISTER_FILE"; then
    echo "! Expected correct content"
    cat "$REGISTER_FILE"
    exit 3
fi

mkdir -p "$GH_TEST_TMP/test116.3" &&
    cd "$GH_TEST_TMP/test116.3" &&
    git init &&
    git config --local githooks.maintainedHooks "all" || exit 1

# Should not have registered in repo 3
if ! grep -qE '.+/test *116.1/.git$' "$REGISTER_FILE" ||
    ! grep -qE '.+/test *116.2/.git$' "$REGISTER_FILE"; then
    echo "! Expected correct content"
    cat "$REGISTER_FILE"
    exit 3
fi

echo "- Install into repo 1,2,3 ..."
echo "Y
$GH_TEST_TMP
" | "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --stdin || exit 1

if ! grep -qE ".+/test *116.1/.git$" "$REGISTER_FILE" ||
    ! grep -qE ".+/test *116.2/.git$" "$REGISTER_FILE" ||
    ! grep -qE ".+/test *116.3/.git$" "$REGISTER_FILE"; then
    echo "! Expected all repos to be registered"
    cat "$REGISTER_FILE"
    exit 4
fi

# Test uninstall to only repo 1
cd "$GH_TEST_TMP/test116.1" || exit 1
if ! "$GH_TEST_BIN/githooks-cli" uninstall; then
    echo "! Uninstall from current repo failed"
    exit 1
fi

if [ "$(git config --local githooks.registered)" = "true" ]; then
    echo "! Expected repo 1 to be marked unregistered"
    exit 1
fi

if grep -qE ".+/test *116.1/.git$" "$REGISTER_FILE" ||
    (! grep -qE ".+/test *116.2/.git$" "$REGISTER_FILE" &&
        ! grep -qE ".+/test *116.3/.git$" "$REGISTER_FILE"); then
    echo "! Expected repo 2 and 3 to still be registered"
    cat "$REGISTER_FILE"
    exit 5
fi

# Test total uninstall to all repos
echo "- Total uninstall..."
echo "Y
$GH_TEST_TMP
" | "$GH_TEST_BIN/githooks-cli" uninstaller --stdin || exit 1

if [ -f "$REGISTER_FILE" ]; then
    echo "! Expected registered list to not exist"
    exit 1
fi

if [ -f "$GH_INSTALL_BIN_DIR/githooks-runner" ] ||
    [ -f "$GH_INSTALL_BIN_DIR/githooks-cli" ]; then
    echo "! Expected that all binaries are deleted."
    exit 1
fi

# Reinstall everywhere
echo "- Reinstall everywhere..."
echo "y

Y
y
$GH_TEST_TMP
" | "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --stdin || exit 1

# Update Test
# Set all other hooks to dirty by adding something
# shellcheck disable=SC2156
find "$GH_TEST_TMP" -type f -path "*/.git/hooks/*" -exec \
    sh -c "echo 'Add DIRTY to {}' && echo '#DIRTY' >>'{}'" \; || exit 1
find "$GH_TEST_TMP" -type f -path "*/.git/hooks/*" |
    while read -r HOOK; do
        if ! grep -q "#DIRTY" "$HOOK"; then
            echo "! Expected hooks to be dirty"
            exit 1
        fi
    done || exit 1

# Trigger the update only from repo 3
CURRENT_TIME=$(date +%s)
MOCK_LAST_RUN=$((CURRENT_TIME - 100000))

# Reset to trigger update from repo 3
if ! git -C "$GH_TEST_REPO" reset --hard v9.9.1 >/dev/null; then
    echo "! Could not reset master to trigger update."
    exit 1
fi

cd "$GH_TEST_TMP/test116.3" &&
    git config --global githooks.updateCheckEnabled true &&
    set_update_check_timestamp $MOCK_LAST_RUN &&
    OUT=$(git commit --allow-empty -m 'Second commit' 2>&1) || exit 1

if ! echo "$OUT" | grep -q "There is a new Githooks update available"; then
    echo "! Expected update-check output not found"
    echo "$OUT"
    exit 1
fi

OUT=$("$GH_INSTALL_BIN_DIR/githooks-cli" update 2>&1)

if ! echo "$OUT" | grep -q "All done! Enjoy!"; then
    echo "! Expected installation output not found"
    echo "$OUT"
    exit 1
fi

# Check that all hooks are updated
find "$GH_TEST_TMP" -type f -path "*/.git/hooks/*" \
    -and -not -name "*disabled*" \
    -and -not -path "*githooks-tmp*" |
    while read -r HOOK; do
        if grep -q "#DIRTY" "$HOOK" && ! echo "$HOOK" | grep -q ".4"; then
            echo "! Expected hooks to be updated $HOOK"
            exit 1
        fi
    done || exit 1

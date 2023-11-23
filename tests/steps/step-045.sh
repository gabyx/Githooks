#!/usr/bin/env bash
# shellcheck disable=SC1091
# Test:
#   Run an install, skipping the intro README files

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "use-core-hookspath"; then
    echo "Using core.hooksPath"
    exit 249
fi

mkdir -p "$GH_TEST_TMP/test045/001" &&
    cd "$GH_TEST_TMP/test045/001" &&
    git init || exit 1
mkdir -p "$GH_TEST_TMP/test045/002" &&
    cd "$GH_TEST_TMP/test045/002" &&
    git init || exit 1

cd "$GH_TEST_TMP/test045" || exit 1

echo "n
y
$GH_TEST_TMP/test045
s
" | "$GH_TEST_BIN/cli" installer --stdin || exit 1

if ! grep -q "github.com/gabyx/githooks" "$GH_TEST_TMP/test045/001/.git/hooks/pre-commit"; then
    echo "! Hooks were not installed into 001"
    exit 1
fi

if [ -f "$GH_TEST_TMP/test045/001/.githooks/README.md" ]; then
    echo "! README was unexpectedly installed into 001"
    exit 1
fi

if ! grep -q "github.com/gabyx/githooks" "$GH_TEST_TMP/test045/002/.git/hooks/pre-commit"; then
    echo "! Hooks were not installed into 002"
    exit 1
fi

if [ -f "$GH_TEST_TMP/test045/002/.githooks/README.md" ]; then
    echo "! README was unexpectedly installed into 002"
    exit 1
fi

# Reset to trigger update from repo 1
# Auto-update should not install Readme.
CURRENT_TIME=$(date +%s)
MOCK_LAST_RUN=$((CURRENT_TIME - 100000))
# shellcheck disable=SC2015
git -C "$GH_TEST_REPO" reset --hard v9.9.1 >/dev/null || {
    echo "! Could not reset server to trigger update."
    exit 1
}

CURRENT="$(git -C ~/.githooks/release rev-parse HEAD)"
cd "$GH_TEST_TMP/test045/001" &&
    git config --global githooks.autoUpdateEnabled true &&
    setUpdateCheckTimestamp $MOCK_LAST_RUN &&
    OUT=$(git commit --allow-empty -m 'Second commit' 2>&1) || exit 1

AFTER="$(git -C ~/.githooks/release rev-parse HEAD)"
if [ "$CURRENT" = "$AFTER" ] ||
    [ "$(git -C "$GH_TEST_REPO" rev-parse v9.9.1)" != "$AFTER" ]; then
    echo "! Release clone was not updated, but it should have!"
    echo "$OUT"
    exit 1
fi

if ! echo "$OUT" | grep -E "Installed.*into.*/002"; then
    echo "! Auto update should have installed into registered repo 2"
    echo "$OUT"
    exit 1
fi

# Check that no Readme is installed.
if [ -f "$GH_TEST_TMP/test045/001/.githooks/README.md" ] ||
    [ -f "$GH_TEST_TMP/test045/002/.githooks/README.md" ]; then
    echo "! README was unexpectedly installed into 001/002 during auto-update"
    exit 1
fi

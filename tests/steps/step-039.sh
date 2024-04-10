#!/usr/bin/env bash
# Test:
#   Run a install successfully and install run wrappers into the current repo.

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

mkdir -p "$GH_TEST_TMP/start/dir" && cd "$GH_TEST_TMP/start/dir" || exit 1

# Place no hooks into the repository
mkdir -p "$GH_TEST_TMP/empty" &&
    GIT_TEMPLATE_DIR="$GH_TEST_TMP/empty" \
        git init || exit 1

if ! "$GH_TEST_BIN/githooks-cli" installer; then
    echo "! Installation failed"
    exit 1
fi

path=$(git config --global githooks.pathForUseCoreHooksPath)
[ -d "$path" ] || {
    echo "! Path '$path' does not exist."
    exit 1
}

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    OUT=$("$GH_TEST_BIN/githooks-cli" install 2>&1)
    # shellcheck disable=SC2181
    if [ $? -eq 0 ] || ! echo "$OUT" | grep -q "has no effect"; then
        echo "! Install into current should have failed, because using 'core.hooksPath'"
        echo "$OUT"
        exit 1
    fi
else
    if ! "$GH_TEST_BIN/githooks-cli" install; then
        echo "! Install into current repo should have succeeded"
        exit 1
    fi

    if [ "$path" != "$(git config --local core.hooksPath)" ]; then
        echo "! Config 'core.hooksPath' does not point to the same directory."
        git config --local core.hooksPath
        exit 1
    fi
fi

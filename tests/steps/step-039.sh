#!/usr/bin/env bash
# Test:
#   Run a install successfully and install run wrappers into the current repo.

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

# Place no hooks into the repository
mkdir -p "$GH_TEST_TMP/start/dir" && cd "$GH_TEST_TMP/start/dir" &&
    git init --template=/dev/null || exit 1

if ! "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}"; then
    echo "! Installation failed"
    exit 1
fi

if is_centralized_tests; then
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

    check_local_install "$GH_TEST_TMP/start/dir"
fi

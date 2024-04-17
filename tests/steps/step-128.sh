#!/usr/bin/env bash
# Test:
#   Clone repo with submodule which contains hooks.

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

"$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" || exit 1
# Do not accept trust prompt.
# accept_all_trust_prompts || exit 1

# Make submodule with post-checkout hook.
mkdir -p "$GH_TEST_TMP/test128-sub/.githooks" &&
    cd "$GH_TEST_TMP/test128-sub" &&
    echo "echo \"Hello: \$(pwd)\" > '$GH_TEST_TMP/test128.out'" >".githooks/post-checkout" &&
    git init &&
    git add . &&
    git commit --allow-empty -m "Init" || exit 1

# Make normal repo with submodule.
echo "Make normal repo with submodule"
mkdir -p "$GH_TEST_TMP/test128" &&
    cd "$GH_TEST_TMP/test128" &&
    git init &&
    GITHOOKS_SKIP_UNTRUSTED_HOOKS=true \
        git submodule add "file://$GH_TEST_TMP/test128-sub" sub &&
    git add . &&
    git commit -a -m "Submodule added" || exit 1

# Clone project with submodule
echo "Clone project"
git clone "$GH_TEST_TMP/test128" "$GH_TEST_TMP/test128-clone" &>/dev/null &&
    cd "$GH_TEST_TMP/test128-clone" || exit 1

# Init submodule
out=$(GITHOOKS_SKIP_UNTRUSTED_HOOKS=true \
    git submodule update --init --recursive 2>&1)

# shellcheck disable=SC2181
if [ $? -ne 0 ] || grep -q "Hello:" "$GH_TEST_TMP/test128.out"; then
    echo "$out"
    echo "! Submodule init/updated and hook should not have run run."
    exit 1
fi

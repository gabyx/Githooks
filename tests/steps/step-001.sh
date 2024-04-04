#!/usr/bin/env bash
# Test:
#   Run a simple install non-interactively and verify the hooks are in place

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

# run the default install
"$GH_TEST_BIN/githooks-cli" installer --non-interactive || exit 1

mkdir -p "$GH_TEST_TMP/test1" &&
    cd "$GH_TEST_TMP/test1" &&
    git init || exit 1

# verify that the pre-commit is installed

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "use-core-hookspath"; then
    grep -q 'https://github.com/gabyx/githooks' "$(git config core.hooksPath)/pre-commit"
else
    grep -q 'https://github.com/gabyx/githooks' .git/hooks/pre-commit
fi

#!/usr/bin/env bash
# Test:
#   Run an install, and let it search for the template dir

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step

accept_all_trust_prompts || exit 1

# move the built-in git template folder
mkdir -p "$GH_TEST_TMP/git-templates" &&
    mv "$GH_TEST_GIT_CORE/templates" "$GH_TEST_TMP/git-templates/" &&
    rm -f "$GH_TEST_TMP/git-templates/templates/hooks/"* &&
    touch "$GH_TEST_TMP/git-templates/templates/hooks/pre-commit.sample" ||
    exit 1

export GIT_TEMPLATE_DIR="$GH_TEST_TMP/git-templates/templates"

# run the install, and let it search for the templates
OUT=$("$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" --hooks-dir-use-template-dir 2>&1)
EXIT_CODE="$?"

if is_centralized_tests; then
    if [ "$EXIT_CODE" -eq "0" ] || ! echo "$OUT" |
        grep -C 10 "You cannot use 'centralized'" |
        grep -C 10 "duplicating run-wrappers" |
        grep -q "is nonsense"; then
        echo "! Expected install to fail."
        echo "$OUT"
        exit 1
    fi

    # Further test are not useful for centralized install.
    exit 0
else
    if [ "$EXIT_CODE" -ne "0" ]; then
        echo "! Expected install to succeed."
        exit 1
    fi
fi

check_install

if ! [ -f "$GH_TEST_TMP/git-templates/templates/hooks/pre-commit" ]; then
    echo "! Expected hook is not installed"
    exit 1
fi

if ! grep 'github.com/gabyx/githooks' "$GH_TEST_TMP/git-templates/templates/hooks/pre-commit"; then
    echo "! Expected hook doesn't have the expected contents"
    exit 1
fi

mkdir -p "$GH_TEST_TMP/test6" &&
    cd "$GH_TEST_TMP/test6" &&
    git init || exit 1

check_local_install_run_wrappers

# Reinstall and check again.
"$GH_INSTALL_BIN_DIR/githooks-cli" install
check_local_install_run_wrappers

#!/usr/bin/env bash
# Test:
#   Set up local repos, run the install and skip installing hooks into existing directories

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    echo "Using centralized install"
    exit 249
fi

mkdir -p ~/test100/p001 ~/test100/p002 &&
    cd ~/test100/p001 &&
    git init &&
    cd ~/test100/p002 &&
    git init || exit 1

if grep -r 'github.com/gabyx/githooks' ~/test100/; then
    echo "! Hooks were installed ahead of time"
    exit 1
fi

# run the install, and skip installing the hooks into existing repos
echo 'y

n
' | "$GH_TEST_BIN/githooks-cli" installer --stdin --skip-install-into-existing || exit 1

check_no_local_install ~/test100/p001
check_no_local_install ~/test100/p002

# run the install again, and let it install into existing repos
echo 'n
y
' | "$GH_TEST_BIN/githooks-cli" installer --stdin

check_local_install ~/test100/p001
check_local_install ~/test100/p002

rm -rf ~/test100

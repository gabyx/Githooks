#!/usr/bin/env bash
# Test:
#   Set up local repos, run the install and verify the hooks get installed (home directory)

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "centralized"; then
    echo "Using centralized install"
    exit 249
fi

mkdir -p ~/test021/p001 &&
    cd ~/test021/p001 &&
    git init || exit 1

mkdir -p ~/test021/p002 &&
    cd ~/test021/p002 &&
    git init || exit 1

# run the install, and select installing the hooks into existing repos
echo 'n
y
~/test021
' | "$GH_TEST_BIN/githooks-cli" installer --stdin || exit 1

check_local_install ~/test021/p001
check_local_install ~/test021/p002

rm -rf ~/test021

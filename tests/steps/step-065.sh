#!/usr/bin/env bash
# Test:
#   Run an install without an existing template directory and refusing to set a new one up

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

if echo "${EXTRA_INSTALL_ARGS:-}" | grep -q "use-core-hookspath"; then
    echo "Using core.hooksPath"
    exit 249
fi

accept_all_trust_prompts || exit 1

rm -rf "$GH_TEST_GIT_CORE/templates/hooks"

echo 'n
' | "$GH_TEST_BIN/cli" installer --stdin

# shellcheck disable=SC2181
if [ $? -eq 0 ]; then
    echo "! Expected to fail"
    exit 1
fi

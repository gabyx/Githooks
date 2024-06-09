#!/usr/bin/env bash
# Test:
#   Shared hooks `prepare-commit-msg` should pass commit message file.
#   Issue #172.

set -e
set -u

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

init_step
accept_all_trust_prompts || exit 1

git config --global githooks.testingTreatFileProtocolAsRemote "true"

"$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}" || exit 1

if ! "$GH_TEST_BIN/githooks-cli" installer "${EXTRA_INSTALL_ARGS[@]}"; then
    echo "! Failed to execute the install script"
    exit 1
fi

# Make shared repository.
mkdir -p "$GH_TEST_TMP/shared/test-148.git/.githooks/prepare-commit-msg"
cat <<'EOF' >"$GH_TEST_TMP/shared/test-148.git/.githooks/prepare-commit-msg/test.sh"
#!/bin/sh
echo "Args:" "$@"
echo "Msg: '$(cat <"$1")'"
EOF

cd "$GH_TEST_TMP/shared/test-148.git" &&
    git init &&
    chmod +x .githooks/prepare-commit-msg/test.sh &&
    git add . &&
    git commit -m 'Testing' || exit 1

mkdir -p "$GH_TEST_TMP/test148" &&
    cd "$GH_TEST_TMP/test148" &&
    mkdir -p .githooks &&
    echo -e "urls:\n  - file://$GH_TEST_TMP/shared/test-148.git" >.githooks/.shared.yaml &&
    git init || exit 2

"$GH_TEST_BIN/githooks-cli" install
"$GH_TEST_BIN/githooks-cli" shared update

echo "Make test commit"
OUT=$(git commit --allow-empty -m "fix: Test prepare-commit-msg args." 2>&1)

if ! echo "$OUT" | grep "Args: .git/COMMIT_EDITMSG message" ||
    ! echo "$OUT" | grep "Msg: 'fix: Test prepare-commit-msg args.'"; then
    echo "Could not get prepare-commit-message" >&2
    echo "$OUT"
    exit 1
fi

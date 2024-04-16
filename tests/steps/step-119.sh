#!/usr/bin/env bash
# Test:
#   Direct runner execution: execute in parallel and check priority list and sequence.

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

accept_all_trust_prompts || exit 1

"$GH_TEST_BIN/githooks-cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/shared-119.git/githooks/pre-commit" &&
    cd "$GH_TEST_TMP/shared-119.git" &&
    echo 'echo "-shared-step 0.1"' >"githooks/pre-commit/step-0" &&
    mkdir -p githooks/pre-commit/step-1 &&
    echo 'echo "-shared-step 1.1"' >"githooks/pre-commit/step-1/step-1.1" &&
    echo 'echo "-shared-step 1.2"' >"githooks/pre-commit/step-1/step-1.2" &&
    mkdir -p githooks/pre-commit/step-2 &&
    echo 'echo "-shared-step 2.1"' >"githooks/pre-commit/step-2/step-2.1" &&
    echo 'echo "-shared-step 2.2"' >"githooks/pre-commit/step-2/step-2.2" &&
    echo 'echo "-shared-step 3.1"' >"githooks/pre-commit/step-3" &&
    git init &&
    git add . &&
    git commit --no-verify -m 'Initial commit' ||
    exit 2

mkdir -p "$GH_TEST_TMP/test119/.githooks/pre-commit" &&
    cd "$GH_TEST_TMP/test119" &&
    echo 'echo "-step 10.1"' >".githooks/pre-commit/step-10" &&
    mkdir -p .githooks/pre-commit/step-11 &&
    echo 'echo "-step 11.1"' >".githooks/pre-commit/step-11/step-11.1" &&
    echo 'echo "-step 11.2"' >".githooks/pre-commit/step-11/step-11.2" &&
    mkdir -p .githooks/pre-commit/step-12 &&
    echo 'echo "-step 12.1"' >".githooks/pre-commit/step-12/step-12.1" &&
    echo 'echo "-step 12.2"' >".githooks/pre-commit/step-12/step-12.2" &&
    echo 'echo "-step 13.1"' >".githooks/pre-commit/step-13" &&
    git init &&
    install_hooks_if_not_centralized &&
    git add . &&
    git commit --no-verify -m 'Initial commit' ||
    exit 3

"$GH_INSTALL_BIN_DIR/githooks-cli" shared add --local "file://$GH_TEST_TMP/shared-119.git" || exit 4
"$GH_INSTALL_BIN_DIR/githooks-cli" shared update || exit 5

OUT=$(git commit --allow-empty -m "test" 2>&1)

# Check execution prio list
SEQ_FILE=$(echo "$OUT" | grep -m 1 "prio-list-pre-commit" | sed -E "s/.*'(.*)'.*/\1/")
if [ ! -f "$SEQ_FILE" ]; then
    echo "! Prio list '$SEQ_FILE' does not exist"
    exit 6
fi

jq -e '.LocalHooks[0][0].BatchName == "step-10"' "$SEQ_FILE" || exit 7
jq -e '.LocalHooks[1][0].BatchName == "step-11"' "$SEQ_FILE" || exit 9
jq -e '.LocalHooks[1][1].BatchName == "step-11"' "$SEQ_FILE" || exit 10
jq -e '.LocalHooks[2][0].BatchName == "step-12"' "$SEQ_FILE" || exit 12
jq -e '.LocalHooks[2][1].BatchName == "step-12"' "$SEQ_FILE" || exit 13
jq -e '.LocalHooks[3][0].BatchName == "step-13"' "$SEQ_FILE" || exit 15
jq -e '.LocalHooks[0] | length == 1' "$SEQ_FILE" || exit 8
jq -e '.LocalHooks[1] | length == 2' "$SEQ_FILE" || exit 11
jq -e '.LocalHooks[2] | length == 2' "$SEQ_FILE" || exit 14
jq -e '.LocalHooks[3] | length == 1' "$SEQ_FILE" || exit 16

jq -e '.LocalSharedHooks[0][0].BatchName == "step-0"' "$SEQ_FILE" || exit 20
jq -e '.LocalSharedHooks[1][0].BatchName == "step-1"' "$SEQ_FILE" || exit 22
jq -e '.LocalSharedHooks[1][1].BatchName == "step-1"' "$SEQ_FILE" || exit 23
jq -e '.LocalSharedHooks[2][0].BatchName == "step-2"' "$SEQ_FILE" || exit 25
jq -e '.LocalSharedHooks[2][1].BatchName == "step-2"' "$SEQ_FILE" || exit 26
jq -e '.LocalSharedHooks[3][0].BatchName == "step-3"' "$SEQ_FILE" || exit 28
jq -e '.LocalSharedHooks[0] | length == 1' "$SEQ_FILE" || exit 21
jq -e '.LocalSharedHooks[1] | length == 2' "$SEQ_FILE" || exit 24
jq -e '.LocalSharedHooks[2] | length == 2' "$SEQ_FILE" || exit 27
jq -e '.LocalSharedHooks[3] | length == 1' "$SEQ_FILE" || exit 29

# Test the sequence in the output
SEQ1="-step 10.1;-step 11.1;-step 11.2;-step 12.1;-step 12.2;-step 13.1;"
SEQ2="-shared-step 0.1;-shared-step 1.1;-shared-step 1.2;-shared-step 2.1;-shared-step 2.2;-shared-step 3.1;"

if ! echo "$OUT" | tr '\r\n' ';' | grep -q -F -e "$SEQ1"; then
    echo "! Execution sequence 1 not found"
    echo "$OUT" | tr '\r\n' ';'
    exit 30
fi

if ! echo "$OUT" | tr '\r\n' ';' | grep -q -F -e "$SEQ2"; then
    echo "! Execution sequence 2 not found"
    echo "$OUT" | tr '\r\n' ';'
    exit 31
fi

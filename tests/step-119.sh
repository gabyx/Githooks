#!/bin/sh
# Test:
#   Direct runner execution: execute in parallel and check sequence.

"$GH_TEST_BIN/cli" installer || exit 1

mkdir -p "$GH_TEST_TMP/shared-119.git/githooks/pre-commit" &&
    cd "$GH_TEST_TMP/shared-119.git" &&
    echo 'echo "-shared-step 1.1"' >"githooks/pre-commit/step-1" &&
    mkdir -p githooks/pre-commit/step-2 &&
    echo 'echo "-shared-step 2.1"' >"githooks/pre-commit/step-2/step-2.1" &&
    echo 'echo "-shared-step 2.2"' >"githooks/pre-commit/step-2/step-2.2" &&
    mkdir -p githooks/pre-commit/step-3 &&
    echo 'echo "-shared-step 3.1"' >"githooks/pre-commit/step-3/step-3.1" &&
    echo 'echo "-shared-step 3.2"' >"githooks/pre-commit/step-3/step-3.2" &&
    echo 'echo "-shared-step 4.1"' >"githooks/pre-commit/step-4" &&
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
    git add . &&
    git commit --no-verify -m 'Initial commit' ||
    exit 3

git hooks shared add --local "file://$GH_TEST_TMP/shared-119.git" || exit 4
git hooks shared update || exit 5

OUT=$(git commit --allow-empty -m "test" 2>&1)

SEQ1="-step 10.1;-step 11.1;-step 11.2;-step 12.1;-step 12.2;-step 13.1;"
SEQ2="-shared-step 1.1;-shared-step 2.1;-shared-step 2.2;-shared-step 3.1;-shared-step 3.2;-shared-step 4.1;"

if ! echo "$OUT" | tr '\r\n' ';' | grep -q -F -e "$SEQ1"; then
    echo "! Execution sequence 1 not found"
    echo "$OUT" | tr '\r\n' ';'
    exit 6
fi

if ! echo "$OUT" | tr '\r\n' ';' | grep -q -F -e "$SEQ2"; then
    echo "! Execution sequence 2 not found"
    echo "$OUT" | tr '\r\n' ';'
    exit 7
fi

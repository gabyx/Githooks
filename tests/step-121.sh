#!/usr/bin/env bash
# Test:
#   Direct runner execution: test a single pre-commit hook file with a runner script

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

[ "$(id -u)" -eq 0 ] && ROOT_ACCESS="true"

function cleanup() {
    git config --unset --global "githooks.monkey"
    [ -n "$ROOT_ACCESS" ] && git config --unset --system "githooks.monkey"
}

trap cleanup EXIT

mkdir -p "$GH_TEST_TMP/test121" &&
    cd "$GH_TEST_TMP/test121" &&
    git init || exit 3

# Make our own runner.

cat <<"EOF" >"custom-runner.go" || exit 3
package main

import (
    "fmt"
    "os"
    "strings"
)

func main() {
    fmt.Printf("Hello\n")
    fmt.Printf("File:%s\n", os.Args[1])
    fmt.Printf("Args:%s\n", strings.Join(os.Args[2:], ","))
    fmt.Printf("Env:%s\n", os.Getenv("MY_ENV_VAR_A"))
    fmt.Printf("Env:%s\n", os.Getenv("MY_ENV_VAR_B"))
}
EOF

go build -o custom-runner.exe custom-runner.go || exit 4

git config --global 'githooks.monkey' "git-monkey-global"
[ -n "$ROOT_ACCESS" ] &&
    SYSTEM_VALUE="git-monkey-system" &&
    git config --system 'githooks.monkey' "$SYSTEM_VALUE"

# shellcheck disable=SC2016
mkdir -p .githooks &&
    cat <<"EOF" >".githooks/pre-commit.yaml" || exit 5
cmd: "${env:GH_TEST_TMP}/test121/custom-runner.exe"
args:
    - "my-file.py"
    - '${env:MONKEY}'
    - "\${env:MONKEY}"
    - "${git:githooks.monkey}"
    - "${git-g:githooks.monkey}"
    - "${git-s:githooks.monkey}"
env:
    - "MY_ENV_VAR_A=${env:MONKEY}-A"
    - "MY_ENV_VAR_B=${env:MONKEY}-B"
version: 1
EOF

# Execute pre-commit by the runner
OUT=$(MONKEY="mon key" "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit 2>&1)
# shellcheck disable=SC2181,SC2016
if [ "$?" -ne 0 ] ||
    ! echo "$OUT" | grep "Hello" ||
    ! echo "$OUT" | grep "my-file.py" ||
    ! echo "$OUT" | grep "Args:mon key,\${env:MONKEY},git-monkey-global,git-monkey-global,$SYSTEM_VALUE" ||
    ! echo "$OUT" | grep "Env:mon key-A" ||
    ! echo "$OUT" | grep "Env:mon key-B"; then
    echo "! Expected hook with runner command to be executed."
    echo "$OUT"
    exit 6
fi

# Add local git config
git config --local 'githooks.monkey' "git-monkey"

mkdir -p .githooks &&
    cat <<"EOF" >".githooks/pre-commit.yaml" || exit 5
cmd: "${env:GH_TEST_TMP}/test121/custom-runner.exe"
args:
    - "my-file.py"
    - '${env:MONKEY}'
    - "\${env:MONKEY}"
    - "${git:githooks.monkey}"
    - "${git-l:githooks.monkey}"
    - "${git-g:githooks.monkey}"
    - "${git-s:githooks.monkey}"
version: 1
EOF

# Execute pre-commit by the runner
OUT=$(MONKEY="mon key" "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit 2>&1)
# shellcheck disable=SC2181,SC2016
if [ "$?" -ne 0 ] ||
    ! echo "$OUT" | grep "Args:mon key,\${env:MONKEY},git-monkey,git-monkey,git-monkey-global,$SYSTEM_VALUE"; then
    echo "! Expected hook with runner command to be executed."
    echo "$OUT"
    exit 6
fi

# Testing "!" operator.
# Test if it fails!
cat <<"EOF" >".githooks/pre-commit.yaml" || exit 5
cmd: "${env:GH_TEST_TMP}/test121/custom-runner.exe"
args:
    - "my-file.py"
    - "${!env:MONKEY}"
    - "\${env:MONKEY}"
version: 1
EOF

OUT=$("$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit 2>&1)
# shellcheck disable=SC2181,SC2016
if [ "$?" -eq 0 ] || ! echo "$OUT" | grep "Error in hook run config"; then
    echo "! Expected hook to fail."
    exit 7
fi

# Testing GITHOOKS_OS/GITHOOKS_ARCH
# Test if it does not fail!
cat <<"EOF" >".githooks/pre-commit.yaml" || exit 5
cmd: "${env:GH_TEST_TMP}/test121/custom-runner.exe"
args:
    - "my-file.py"
    - "${!env:GITHOOKS_OS}"
    - "${!env:GITHOOKS_ARCH}"
version: 1
EOF

OUT=$("$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit 2>&1)
# shellcheck disable=SC2181,SC2016
if [ "$?" -ne 0 ] || echo "$OUT" | grep "Error in hook run config"; then
    echo "! Expected hook to succeed."
    exit 8
fi

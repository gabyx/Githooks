#!/bin/sh
# Test:
#   Direct runner execution: test a single pre-commit hook file with a runner script

mkdir -p "$GH_TEST_TMP/test121" &&
    cd "$GH_TEST_TMP/test121" || exit 2
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
}
EOF

go build -o custom-runner.exe ./... || exit 4

git config --local 'githooks.monkey' "git-monkey"

# shellcheck disable=SC2016
mkdir -p .githooks &&
    cat <<"EOF" >".githooks/pre-commit.yaml" || exit 5
cmd: "${env:GH_TEST_TMP}/test121/custom-runner.exe"
args:
    - "my-file.py"
    - '${env:MONKEY}'
    - "\${env:MONKEY}"
    - "${git-l:githooks.monkey}"
    - "${git:githooks.monkey}"
version: 1
EOF

# Execute pre-commit by the runner
OUT=$(MONKEY="mon key" "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit 2>&1)

# shellcheck disable=SC2181,SC2016
if [ "$?" -ne 0 ] ||
    ! echo "$OUT" | grep "Hello" ||
    ! echo "$OUT" | grep "my-file.py" ||
    ! echo "$OUT" | grep 'Args:mon key,${env:MONKEY},git-monkey,git-monkey'; then
    echo "! Expected hook with runner command to be executed."
    echo "$OUT"
    exit 6
fi

# Test if it fails!
cat <<"EOF" >".githooks/pre-commit.yaml" || exit 5
cmd: "${env:GH_TEST_TMP}/test121/custom-runner.exe"
args:
    - "my-file.py"
    - '${!env:MONKEY}'
    - "\${env:MONKEY}"
version: 1
EOF

OUT=$("$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit 2>&1)
# shellcheck disable=SC2181,SC2016
if [ "$?" -eq 0 ] || ! echo "$OUT" | grep "Error in hook run config"; then
    echo "! Expected hook to fail."
    exit 7
fi

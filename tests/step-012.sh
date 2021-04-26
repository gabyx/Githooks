#!/bin/sh
# Test:
#   Direct runner execution: test a single pre-commit hook file

TEST_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1090
. "$TEST_DIR/general.sh"

acceptAllTrustPrompts || exit 1

mkdir -p "$GH_TEST_TMP/test12" &&
    cd "$GH_TEST_TMP/test12" &&
    git init || exit 1

mkdir -p .githooks &&
    echo "echo 'Direct execution' > '$GH_TEST_TMP/test012.out'" >.githooks/pre-commit &&
    echo "echo \"\$GITHOOKS_OS\" > '$GH_TEST_TMP/test012env.out'" >>.githooks/pre-commit &&
    echo "echo \"\$GITHOOKS_ARCH\" >> '$GH_TEST_TMP/test012env.out'" >>.githooks/pre-commit &&
    "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit ||
    exit 1

# From https://github.com/golang/go/blob/master/src/go/build/syslist.go
goosList="aix|android|darwin|dragonfly|freebsd|hurd|illumos|ios|js|linux|nacl|netbsd|openbsd|plan9|solaris|windows|zos"
goarchList="386|amd64|amd64p32|arm|armbe|arm64|arm64be|ppc64|ppc64le|mips|mipsle|mips64|mips64le|mips64p32|mips64p32le|ppc|riscv|riscv64|s390|s390x|sparc|sparc64|wasm"

if ! grep -q 'Direct execution' "$GH_TEST_TMP/test012.out" ||
    ! grep -Eq "$goosList" "$GH_TEST_TMP/test012env.out" ||
    ! grep -Eq "$goarchList" "$GH_TEST_TMP/test012env.out"; then
    echo "! Expected GITHOOKS_OS and GITHOOKS_ARCH to be defined."
    exit 4
fi

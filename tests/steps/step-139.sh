#!/usr/bin/env bash
# Test:
#   Direct runner execution: list of staged files in a file

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

mkdir -p "$GH_TEST_TMP/test139/.githooks/pre-commit" &&
    cd "$GH_TEST_TMP/test139" &&
    git init &&
    git config githooks.exportStagedFilesAsFile true ||
    exit 1

echo "Test" >>sample.txt
echo "Test" >>second.txt

cat <<EOF >.githooks/pre-commit/print-changes
#!/bin/bash
echo "STAGED_FILES_FILE: \$STAGED_FILES_FILE"
while read -d $'\\0' line ; do
    echo "staged: \$line" >> "$GH_TEST_TMP/test139.out"
done < "\$STAGED_FILES_FILE"
EOF

chmod +x .githooks/pre-commit/print-changes

git add sample.txt second.txt

ACCEPT_CHANGES=A \
    "$GH_TEST_BIN/runner" "$(pwd)"/.git/hooks/pre-commit

if ! grep 'staged: sample.txt' "$GH_TEST_TMP/test139.out"; then
    echo "! Failed to find expected output (1)"
    exit 1
fi

if ! grep 'staged: second.txt' "$GH_TEST_TMP/test139.out"; then
    echo "! Failed to find expected output (2)"
    exit 1
fi

if grep -vE '(sample|second)\.txt' "$GH_TEST_TMP/test139.out"; then
    echo "! Unexpected additional output"
    exit 1
fi

if [ "$(find "$GH_TEST_TMP/test139" -name "*githooks-staged-files*" | wc -l)" != 0 ]; then
    echo "! File .githooks-staged-files is not deleted!"
    exit 1
fi

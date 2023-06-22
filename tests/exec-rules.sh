#!/usr/bin/env bash
set -e
set -u

if [ "${DOCKER_RUNNING:-}" != "true" ]; then
    echo "! This script is only meant to be run in a Docker container"
    exit 1
fi

DIR="$(cd "$(dirname "$0")" >/dev/null 2>&1 && pwd)"

REPO_DIR="$DIR/.."
GO_SRC="$REPO_DIR/githooks"

cd "$GO_SRC" || exit 1

echo "Go generate ..."
go mod vendor
go generate -mod vendor ./...

cd "$REPO_DIR" || exit 1

FAILURES=""

function run_pre_commit_test() {
    echo "Run pre-commit '$1'..."
    if ! bash "$REPO_DIR/.githooks/pre-commit/$1"; then
        FAILURES="$FAILURES
  - $1 failed"
    fi
}

run_pre_commit_test gofmt
run_pre_commit_test golint
run_pre_commit_test no-tabs
run_pre_commit_test no-todo-or-fixme
run_pre_commit_test no-setx
run_pre_commit_test shfmt
run_pre_commit_test shellcheck
run_pre_commit_test shellcheck-ignore-format
run_pre_commit_test cli-docs-up-to-date

if [ -n "$FAILURES" ]; then
    echo "The following pre-commit checks had problems: $FAILURES"
    exit 1
else
    echo "All pre-commit hooks have been verified"
fi

exit 0

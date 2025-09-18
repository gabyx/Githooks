#!/usr/bin/env bash
# shellcheck disable=SC1091

set -e
set -u

ROOT_DIR=$(git rev-parse --show-toplevel)
TEST_DIR="$ROOT_DIR/tests"
. "$ROOT_DIR/tests/general.sh"

cd "$ROOT_DIR"

cat <<EOF | run_docker build --force-rm -t githooks:alpine-lfs-centralized-base -
FROM golang:1.22-alpine
RUN apk update && apk add git git-lfs
RUN apk add bash jq curl docker

# CVE https://github.blog/2022-10-18-git-security-vulnerabilities-announced/#cve-2022-39253
RUN git config --system protocol.file.allow always
EOF

exec "$TEST_DIR/exec-tests.sh" 'alpine-lfs-centralized' --test-centralized-install "$@"

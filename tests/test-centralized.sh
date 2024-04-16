#!/usr/bin/env bash
# shellcheck disable=SC1091

set -e
set -u

ROOT_DIR=$(git rev-parse --show-toplevel)
TEST_DIR="$ROOT_DIR/tests"
. "$ROOT_DIR/tests/general.sh"

cd "$ROOT_DIR"

cat <<EOF | docker build --force-rm -t githooks:alpine-lfs-centralized-base -
FROM golang:1.20-alpine
RUN apk update && apk add git git-lfs
RUN apk add bash jq curl docker

# CVE https://github.blog/2022-10-18-git-security-vulnerabilities-announced/#cve-2022-39253
RUN git config --system protocol.file.allow always

ENV EXTRA_INSTALL_ARGS --centralized
EOF

exec "$TEST_DIR/exec-tests.sh" 'alpine-lfs-centralized' "$@"
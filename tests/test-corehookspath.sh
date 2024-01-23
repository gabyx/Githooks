#!/usr/bin/env bash

set -e
set -u

TEST_DIR=$(cd "$(dirname "$0")" && pwd)

cat <<EOF | docker build --force-rm -t githooks:alpine-lfs-corehookspath-base -
FROM golang:1.20-alpine
RUN apk update && apk add git git-lfs
RUN apk add bash jq curl docker

# CVE https://github.blog/2022-10-18-git-security-vulnerabilities-announced/#cve-2022-39253
RUN git config --system protocol.file.allow always

ENV EXTRA_INSTALL_ARGS --use-core-hookspath
EOF

exec "$TEST_DIR/exec-tests.sh" 'alpine-lfs-corehookspath' "$@"

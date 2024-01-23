#!/usr/bin/env bash

set -e
set -u

TEST_DIR=$(cd "$(dirname "$0")" && pwd)

cat <<EOF | docker build --force-rm -t githooks:alpine-lfs-whitespace-base -
FROM golang:1.20-alpine
RUN apk update && apk add git git-lfs
RUN apk add bash jq curl docker

# CVE https://github.blog/2022-10-18-git-security-vulnerabilities-announced/#cve-2022-39253
RUN git config --system protocol.file.allow always

RUN mkdir -p "/root/whitespace folder"
ENV HOME="/root/whitespace folder"

EOF

# shellcheck disable=SC2016
export ADDITIONAL_INSTALL_STEPS='
# add a space in paths
RUN find "$GH_TESTS" -name "*.sh" -exec sed -i -E "s|GH_TEST_TMP(\})?/test([0-9.]+)|GH_TEST_TMP\1/test \2|g" {} \;
'

exec "$TEST_DIR/exec-tests.sh" 'alpine-lfs-whitespace' "$@"

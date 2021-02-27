#!/bin/sh

TEST_DIR=$(cd "$(dirname "$0")" && pwd)

cat <<EOF | docker build --force-rm -t githooks:alpine-lfs-corehookspath-base -
FROM golang:1.15.8-alpine
RUN apk add git git-lfs --update-cache --repository http://dl-3.alpinelinux.org/alpine/edge/main --allow-untrusted
RUN apk add bash jq
ENV EXTRA_INSTALL_ARGS --use-core-hookspath
EOF

exec sh "$TEST_DIR"/exec-tests.sh 'alpine-lfs-corehookspath' "$@"

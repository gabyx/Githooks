#!/bin/sh

TEST_DIR=$(cd "$(dirname "$0")" && pwd)

cat <<EOF | docker build --force-rm -t githooks:alpine-lfs-base -
FROM golang:1.16-alpine
RUN apk add git git-lfs --update-cache --repository http://dl-3.alpinelinux.org/alpine/edge/main --allow-untrusted
RUN apk add bash jq curl
EOF

exec sh "$TEST_DIR"/exec-tests.sh 'alpine-lfs' "$@"

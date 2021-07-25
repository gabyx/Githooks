#!/bin/sh

TEST_DIR=$(cd "$(dirname "$0")" && pwd)

cat <<EOF | docker build --force-rm -t githooks:alpine-nolfs-base -
FROM golang:1.16-alpine
RUN apk add git --update-cache --repository http://dl-3.alpinelinux.org/alpine/edge/main --allow-untrusted
RUN apk add bash jq curl
EOF

exec sh "$TEST_DIR"/exec-tests.sh 'alpine-nolfs' "$@"

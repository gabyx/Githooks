#!/usr/bin/env bash

cat <<EOF | docker build --force-rm -t githooks:testsuite -
FROM golang:1.17-alpine
RUN apk add git curl git-lfs --update-cache --repository http://dl-3.alpinelinux.org/alpine/edge/main --allow-untrusted
RUN apk add bash jq

RUN curl -sSfL https://raw.githubusercontent.com/golangci/c/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin v1.35.2

# Git refuses to do stuff in this mounted directory.
RUN git config --global safe.directory /githooks
EOF

if ! docker run --rm -it \
    -v "$(pwd)":/githooks \
    -w /githooks githooks:testsuite \
    tests/exec-testsuite.sh; then

    echo "! Check rules had failures."
    exit 1
fi

exit 0

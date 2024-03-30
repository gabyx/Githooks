#!/usr/bin/env bash

set -e
set -u

function clean_up() {
    # shellcheck disable=SC2317
    docker rmi "githooks:unittests" &>/dev/null || true
}

trap clean_up EXIT

cat <<EOF | docker build --force-rm -t githooks:unittests -
FROM golang:1.20-alpine
RUN apk update && apk add git git-lfs
RUN apk add bash jq curl docker

# CVE https://github.blog/2022-10-18-git-security-vulnerabilities-announced/#cve-2022-39253
RUN git config --system protocol.file.allow always

RUN curl -sSfL https://raw.githubusercontent.com/golangci/c/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin v1.35.2

# Git refuses to do stuff in this mounted directory.
RUN git config --global safe.directory /githooks

ENV DOCKER_RUNNING=true
EOF

if ! docker run --rm -it \
    -v "$(pwd)":/githooks \
    -v "/var/run/docker.sock:/var/run/docker.sock" \
    -w /githooks githooks:unittests \
    tests/exec-unittests.sh; then

    echo "! Check rules had failures."
    exit 1
fi

exit 0

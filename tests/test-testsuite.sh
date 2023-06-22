#!/usr/bin/env bash

set -e
set -u

function cleanUp() {
    # shellcheck disable=SC2317
    docker rmi "githooks:testsuite" &>/dev/null || true
}

trap cleanUp EXIT

cat <<EOF | docker build --force-rm -t githooks:testsuite -
FROM golang:1.20-alpine
RUN apk add git curl git-lfs --update-cache --repository http://dl-cdn.alpinelinux.org/alpine/edge/main --allow-untrusted
RUN apk add bash jq docker

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
    -w /githooks githooks:testsuite \
    tests/exec-testsuite.sh; then

    echo "! Check rules had failures."
    exit 1
fi

exit 0

#!/usr/bin/env bash

set -e
set -u

rootDir=$(git rev-parse --show-toplevel)

cat <<EOF | docker build --force-rm -t githooks:test-rules -
FROM golang:1.20-alpine
RUN apk update && apk add git git-lfs
RUN apk add bash jq curl docker just

RUN git config --global safe.directory /data

# CVE https://github.blog/2022-10-18-git-security-vulnerabilities-announced/#cve-2022-39253
RUN git config --system protocol.file.allow always

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin v1.52.2

RUN git config --global user.email "githook@test.com" && \
    git config --global user.name "Githook Tests" && \
    git config --global init.defaultBranch main && \
    git config --global core.autocrlf false

ENV DOCKER_RUNNING=true
EOF

if
    ! docker run --rm -it \
        -v "$rootDir":/data \
        -v "/var/run/docker.sock:/var/run/docker.sock" \
        -e "GH_SHOW_DIFFS=${GH_SHOW_DIFFS:-false}" \
        -w /data githooks:test-rules tests/exec-rules.sh
then
    echo "! Check rules had failures."
    exit 1
fi

exit 0

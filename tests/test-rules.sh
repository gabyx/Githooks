#!/usr/bin/env bash

cat <<EOF | docker build --force-rm -t githooks:test-rules -
FROM golang:1.17-alpine
RUN apk add git curl git-lfs --update-cache --repository http://dl-3.alpinelinux.org/alpine/edge/main --allow-untrusted
RUN apk add bash jq

RUN git config --global safe.directory /data

# CVE https://github.blog/2022-10-18-git-security-vulnerabilities-announced/#cve-2022-39253
RUN git config --global protocol.file.allow always

RUN curl -fsSL https://github.com/mvdan/sh/releases/download/v3.1.1/shfmt_v3.1.1_linux_amd64 -o /usr/local/bin/shfmt \
    && chmod +x /usr/local/bin/shfmt \
    && shfmt --version
RUN T=$(mktemp); curl -fsSL https://github.com/koalaman/shellcheck/releases/download/v0.7.1/shellcheck-v0.7.1.linux.x86_64.tar.xz -o "\\\$T" \
    && tar -xf "\\\$T" --strip-components=1 -C /usr/local/bin/ \
    && chmod +x /usr/local/bin/shellcheck \
    && shellcheck --version

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin v1.34.1

ENV DOCKER_RUNNING=true
EOF

if ! docker run --rm -it -v "$(pwd)":/data -w /data githooks:test-rules tests/exec-rules.sh; then
    echo "! Check rules had failures."
    exit 1
fi

exit 0

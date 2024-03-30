#!/usr/bin/env bash

set -e
set -u

TEST_DIR=$(cd "$(dirname "$0")" && pwd)

cat <<EOF | docker build \
    --force-rm -t githooks:alpine-user-base -
FROM golang:1.20-alpine
RUN apk update && apk add git git-lfs
RUN apk add bash jq curl

# CVE https://github.blog/2022-10-18-git-security-vulnerabilities-announced/#cve-2022-39253
RUN git config --system protocol.file.allow always
EOF

# shellcheck disable=SC2016,SC1004
export ADDITIONAL_PRE_INSTALL_STEPS='
RUN adduser -D -u 1099 test
RUN if [ -n "$DOCKER_GROUP_ID" ]; then \
        addgroup -g "$DOCKER_GROUP_ID" docker && \
        adduser test docker && \
        apk add docker; \
    else \
        echo "Not adding docker since not working with user!" &>2; \
    fi
RUN rm -rf "$GH_TEST_GIT_CORE/templates/hooks" && \
    mkdir -p "$GH_TEST_REPO" "$GH_TEST_GIT_CORE/templates/hooks" && \
    chown -R test:test "$GH_TEST_REPO" "$GH_TEST_GIT_CORE"
USER test
'

exec "$TEST_DIR/exec-tests.sh" 'alpine-user' "$@"

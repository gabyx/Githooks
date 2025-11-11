#!/usr/bin/env bash
# shellcheck disable=SC1091

set -e
set -u

ROOT_DIR=$(git rev-parse --show-toplevel)
TEST_DIR="$ROOT_DIR/tests"
. "$ROOT_DIR/tests/general.sh"

cd "$ROOT_DIR"

cat <<EOF | docker build \
    --force-rm -t githooks:alpine-user-base -
FROM golang:1.24-alpine
RUN apk update && apk add git git-lfs
RUN apk add bash jq curl

# CVE https://github.blog/2022-10-18-git-security-vulnerabilities-announced/#cve-2022-39253
RUN git config --system protocol.file.allow always
EOF

# shellcheck disable=SC2016,SC1004
export ADDITIONAL_PRE_INSTALL_STEPS='
RUN adduser -D -u 1099 test
RUN if [ -n "$DOCKER_GROUP_ID" ]; then \
    existingGroup=$(getent group "$DOCKER_GROUP_ID" | cut -d: -f1); \
    if [ "$existingGroup" != "" ]; then \
            apk add shadow && \
            newID=$(($DOCKER_GROUP_ID - 1)) && \
            echo "Remapping group id $existingGroup:$DOCKER_GROUP_ID to $newID since existing." && \
            groupmod -g "$newID" "$existingGroup" && \
            apk del shadow; \
        fi; \
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

#!/bin/sh
ROOT_DIR=$(git rev-parse --show-toplevel)
TEST_DIR="$ROOT_DIR/tests"

IMAGE_TYPE="alpine-coverage"

[ -n "$COVERALLS_TOKEN" ] || {
    echo "! You need to set 'COVERALL_TOKEN'."
    exit 1
}

cleanup() {
    docker rmi "githooks:$IMAGE_TYPE-base"
    docker rmi "githooks:$IMAGE_TYPE"
}

trap cleanup EXIT

cat <<EOF | docker build --force-rm -t githooks:$IMAGE_TYPE-base -
FROM golang:1.16-alpine
RUN apk add git git-lfs --update-cache --repository http://dl-3.alpinelinux.org/alpine/edge/main --allow-untrusted
RUN apk add bash jq curl
RUN go get github.com/wadey/gocovmerge
RUN go install github.com/mattn/goveralls@latest

ENV GH_COVERAGE_DIR="/cover"

# Coveralls env. vars for upload.
ENV COVERALLS_TOKEN="$COVERALLS_TOKEN"
EOF

if echo "$IMAGE_TYPE" | grep -q "\-user"; then
    OS_USER="test"
else
    OS_USER="root"
fi

cat <<EOF | docker build --force-rm -t githooks:$IMAGE_TYPE -f - "$ROOT_DIR" || exit 1
FROM githooks:$IMAGE_TYPE-base

ENV GH_TESTS="/var/lib/githooks-tests"
ENV GH_TEST_TMP="/tmp"
ENV GH_TEST_REPO="/var/lib/githooks"
ENV GH_TEST_BIN="/var/lib/githooks/githooks/bin"
ENV GH_TEST_GIT_CORE="/usr/share/git-core"

${ADDITIONAL_PRE_INSTALL_STEPS:-}

# Add sources.
COPY --chown=$OS_USER:$OS_USER githooks "\$GH_TEST_REPO/githooks"
ADD .githooks/README.md \$GH_TEST_REPO/.githooks/README.md
ADD examples "\$GH_TEST_REPO/examples"

# Replace run-wrapper with coverage run-wrapper
RUN cd \$GH_TEST_REPO && \\
    cp githooks/build/embedded/run-wrapper-coverage.sh githooks/build/embedded/run-wrapper.sh

ADD tests/setup-githooks.sh "\$GH_TESTS/"
RUN bash "\$GH_TESTS/setup-githooks.sh" --coverage

ADD tests "\$GH_TESTS"

# Always don't delete LFS Hooks (for testing, default is unset, but cumbersome for tests)
RUN git config --global githooks.deleteDetectedLFSHooks "n"

# Replace some statements which rely on proper CLI output
# The built instrumented executable output test&coverage shit...
RUN sed -i -E 's@cli" shared location(.*)\)@cli" shared location\1 | grep "^/")@g' \\
    "\$GH_TESTS"/step-*

# Replace all runnner/cli/dialog/'git hooks' invocations.
# Foward over 'coverage/forwarder'.
RUN sed -i -E 's@"(.GH_INSTALL_BIN_DIR|.GH_TEST_BIN)/cli"@"\$GH_TEST_REPO/githooks/coverage/forwarder" "\1/cli"@g' \\
    "\$GH_TESTS/exec-steps.sh" \\
    "\$GH_TESTS"/step-*
RUN sed -i -E 's@"(.GH_INSTALL_BIN_DIR|.GH_TEST_BIN)/runner"@"\$GH_TEST_REPO/githooks/coverage/forwarder" "\1/runner"@g' \\
    "\$GH_TESTS"/step-*
RUN sed -i -E 's@"(.GH_INSTALL_BIN_DIR|.GH_TEST_BIN)/dialog"@"\$GH_TEST_REPO/githooks/coverage/forwarder" "\1/dialog"@g' \\
    "\$GH_TESTS"/step-*
RUN sed -i -E 's@".DIALOG"@"\$GH_TEST_REPO/githooks/coverage/forwarder" "\1"@g' \\
    "\$GH_TESTS"/step-*
RUN sed -i -E 's@git hooks@"\$GH_TEST_REPO/githooks/coverage/forwarder" "\$GH_INSTALL_BIN_DIR/cli"@g' \\
    "\$GH_TESTS"/step-*

${ADDITIONAL_INSTALL_STEPS:-}

RUN echo "Git version: \$(git --version)"
WORKDIR \$GH_TESTS
EOF

# Clean all coverage data
if [ -d "$TEST_DIR/cover" ]; then
    rm -rf "$TEST_DIR/cover"/*
fi

# Run the normal tests to add to the coverage
# inside the current repo
docker run --rm \
    -a stdout -a stderr \
    -v "$TEST_DIR/cover":/cover \
    -v "$TEST_DIR/..":/githooks \
    -w /githooks/tests \
    "githooks:$IMAGE_TYPE-base" \
    sh ./exec-testsuite.sh ||
    exit $?

# Run the integration tests
docker run --rm \
    -a stdout -a stderr \
    -v "$TEST_DIR/cover":/cover \
    "githooks:$IMAGE_TYPE" \
    ./exec-steps.sh "$@" || exit $?

# Upload the produced coverage
# inside the current repo
docker run --rm \
    -a stdout -a stderr \
    -v "$TEST_DIR/cover":/cover \
    -v "$TEST_DIR/..":/githooks \
    -w /githooks \
    -e TRAVIS_JOB_ID="$TRAVIS_JOB_ID" \
    -e TRAVIS_JOB_NAME="$TRAVIS_JOB_NAME" \
    -e TRAVIS_JOB_NUMBER="$TRAVIS_JOB_NUMBER" \
    -e TRAVIS_PULL_REQUEST="$TRAVIS_PULL_REQUEST" \
    -e TRAVIS_PULL_REQUEST_BRANCH="$TRAVIS_PULL_REQUEST_BRANCH" \
    -e TRAVIS_PULL_REQUEST_SHA="$TRAVIS_PULL_REQUEST_SHA" \
    "githooks:$IMAGE_TYPE-base" \
    ./tests/upload-coverage.sh || exit $?

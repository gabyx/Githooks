#!/bin/sh

cat <<EOF | docker build --force-rm -t githooks:alpine-coverage-base -
FROM golang:1.15.6-alpine
RUN apk add git git-lfs --update-cache --repository http://dl-3.alpinelinux.org/alpine/edge/main --allow-untrusted
RUN apk add bash

ENV GH_TESTS="/var/lib/githooks-tests"
ENV GH_TEST_TMP="/tmp"
ENV GH_TEST_REPO="/var/lib/githooks"
ENV GH_TEST_BIN="/var/lib/githooks/githooks/bin"
ENV GH_TEST_GIT_CORE="/usr/share/git-core"

${ADDITIONAL_PRE_INSTALL_STEPS:-}

# Add sources
COPY --chown=${OS_USER}:${OS_USER} githooks \${GH_TEST_REPO}/githooks
RUN sed -i -E 's/^bin//' \${GH_TEST_REPO}/githooks/.gitignore # We use the bin folder
ADD .githooks/README.md \${GH_TEST_REPO}/.githooks/README.md
ADD examples \${GH_TEST_REPO}/examples
ADD tests "\$GH_TESTS"

RUN git config --global user.email "githook@test.com" && \\
    git config --global user.name "Githook Tests" && \\
    git config --global init.defaultBranch main && \\
    git config --global core.autocrlf false

RUN echo "Make test gitrepo to clone from ..." && \\
    cd \$GH_TEST_REPO && git init  >/dev/null 2>&1  && \\
    git add . >/dev/null 2>&1  && \\
    git commit -a -m "Before build" >/dev/null 2>&1

# Build binaries for v9.9.0.
#################################
RUN cd \$GH_TEST_REPO/githooks && \\
    git tag "v9.9.0" >/dev/null 2>&1 && \\
     ./scripts/clean.sh && \\
    ./scripts/build.sh --coverage --build-flags "-tags debug,mock,coverage" && \\
    ./bin/cli --version
RUN echo "Commit build v9.9.0 to repo ..." && \\
    cd \$GH_TEST_REPO && \\
    git add . >/dev/null 2>&1 && \\
    git commit -a --allow-empty -m "Version 9.9.0" >/dev/null 2>&1 && \\
    git tag -f "v9.9.0"

# Build binaries for v9.9.1.
#################################
RUN cd \$GH_TEST_REPO/githooks && \\
    git commit -a --allow-empty -m "Before build" >/dev/null 2>&1 && \\
    git tag -f "v9.9.1" && \\
    ./scripts/clean.sh && \\
    ./scripts/build.sh --coverage --build-flags "-tags debug,mock,coverage" && \\
    ./bin/cli --version
RUN echo "Commit build v9.9.1 to repo ..." && \\
    cd \$GH_TEST_REPO && \\
    git commit -a --allow-empty -m "Version 9.9.1" >/dev/null 2>&1 && \\
    git tag -f "v9.9.1"

RUN if [ -n "\$EXTRA_INSTALL_ARGS" ]; then \\
        sed -i -E 's|(.*)/cli\" installer|\1/cli" installer \$EXTRA_INSTALL_ARGS|g' "\$GH_TESTS"/step-* ; \\
    fi

# Always don't delete LFS Hooks (for testing, default is unset, but cumbersome for tests)
RUN git config --global githooks.deleteDetectedLFSHooks "n"

${ADDITIONAL_INSTALL_STEPS:-}

RUN echo "Git version: \$(git --version)"
WORKDIR \$GH_TESTS
EOF

docker run --rm \
    -a stdout -a stderr \
    githooks:"$IMAGE_TYPE" \
    ./exec-steps-go.sh "$@"

RESULT=$?

docker rmi "githooks:$IMAGE_TYPE"
docker rmi "githooks:$IMAGE_TYPE-base"
exit $RESULT

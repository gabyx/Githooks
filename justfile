clean:
  githooks/scripts/clean.sh

build:
  githooks/scripts/build.sh

test *ARGS:
  tests/test-alpine-user.sh {{ARGS}}

testsuite:
  tests/test-testsuite.sh

testrules:
  tests/test-rules.sh

release-test:
  cd githooks && \
    GORELEASER_CURRENT_TAG=v9.9.9 \
    goreleaser release --snapshot --clean --skip-sign --skip-publish --skip-validate

release version:
  scripts/release.sh {{version}}

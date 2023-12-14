clean:
  githooks/scripts/clean.sh

build:
  githooks/scripts/build.sh

testsuite:
  tests/test-testsuite.sh

release-test:
  cd githooks && \
    GORELEASER_CURRENT_TAG=v9.9.9 \
    goreleaser release --snapshot --clean --skip-sign --skip-publish --skip-validate

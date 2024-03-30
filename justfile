set positional-arguments
set shell := ["bash", "-cue"]
root_dir := justfile_directory()

clean:
  cd "{{root_dir}}" && \
    githooks/scripts/clean.sh

build *args:
  cd "{{root_dir}}" && \
    githooks/scripts/build.sh "$@"

test-user *args:
  cd "{{root_dir}}" && \
    tests/test-alpine-user.sh "$@"

test *args:
  cd "{{root_dir}}" && \
    tests/test-alpine.sh "$@"

lint fix="false":
  cd "{{root_dir}}" && \
    GH_FIX="{{fix}}" \
    tests/test-lint.sh

lint-fix: (lint "true")

testsuite:
  cd "{{root_dir}}" && \
    tests/test-testsuite.sh

testrules:
  cd "{{root_dir}}" && \
    tests/test-lint.sh

release-test *args:
  cd "{{root_dir}}/githooks" && \
    GORELEASER_CURRENT_TAG=v9.9.9 \
    goreleaser release --snapshot --clean --skip=sign --skip=publish --skip=validate "$@"

release version update-info="":
  cd "{{root_dir}}" && \
    scripts/release.sh "{{version}}" "{{update-info}}"

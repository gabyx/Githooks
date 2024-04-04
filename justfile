set positional-arguments
set shell := ["bash", "-cue"]
root_dir := justfile_directory()

clean:
  cd "{{root_dir}}" && \
    githooks/scripts/clean.sh

build *args:
  cd "{{root_dir}}" && \
    githooks/scripts/build.sh "$@"

doc *args:
  cd "{{root_dir}}" && \
    githooks/scripts/build-doc.sh "$@"

test-user *args:
  cd "{{root_dir}}" && \
    tests/test-alpine-user.sh "$@"

test *args:
  cd "{{root_dir}}" && \
    tests/test-alpine.sh "$@"

coverage *args:
  cd "{{root_dir}}" && \
    COVERALLS_TOKEN=non-existing tests/test-coverage.sh "$@"

lint fix="false":
  cd "{{root_dir}}" && \
    GH_FIX="{{fix}}" tests/test-lint.sh || \
      echo "Run 'just lint-fix' to fix the files."

lint-fix: (lint "true")

unittests:
  cd "{{root_dir}}" && \
    tests/test-unittests.sh

release-test *args:
  cd "{{root_dir}}/githooks" && \
    GORELEASER_CURRENT_TAG=v9.9.9 \
    goreleaser release --snapshot --clean --skip=sign --skip=publish --skip=validate "$@"

release version update-info="":
  cd "{{root_dir}}" && \
    scripts/release.sh "{{version}}" "{{update-info}}"

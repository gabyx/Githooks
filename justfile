set positional-arguments
set dotenv-load := true
set shell := ["bash", "-cue"]
root_dir := justfile_directory()

default:
  just --list

# Start a development shell.
[group("general")]
develop:
  nix develop --accept-flake-config --show-trace ".#default"

# Format all files.
[group("general")]
format:
  treefmt

# Clean everything.
[group("general")]
clean:
  githooks/scripts/clean.sh

# Build Githooks.
[group("build")]
build *args:
  githooks/scripts/build.sh "$@"

[group("run")]
run *args: build
  githooks/bin/githooks-cli "$@"

# Build Githooks with Nix.
[group("build")]
build-nix *args:
    nix build -L "./nix#default" -o ./nix/result {{args}}

# Generate all docs.
doc *args:
    githooks/scripts/build-doc.sh "$@"

# List all test.
[group("list-tests")]
list-tests:
  cd "{{root_dir}}/tests/steps" && \
    readarray -t files < <(find . -name "*.sh" -name "step-*" | sort) && \
    for f in "${files[@]}"; do \
      printf " - %s: %s\n" "$f" "$(head -3 "$f" | tail -1)"; \
    done

# Run all integration tests for `alpine-user`.
[group("integration-tests")]
test-user *args:
    tests/test-alpine-user.sh "$@"

# Run all unit tests for `alpine`.
[group("integration-tests")]
test *args:
  cd "{{root_dir}}" && \
    tests/test-alpine.sh "$@"

# Run all coverage.
[group("coverage")]
coverage *args:
  cd "{{root_dir}}" && \
    COVERALLS_TOKEN=non-existing tests/test-coverage.sh "$@"

# Lint everything (local).
[group("lint")]
lint-local:
  cd githooks && \
    golangci-lint run --config "{{root_dir}}/.golangci.yaml" --fix ./...

# Lint everything (dockerized).
[group("lint")]
lint fix="false":
  cd "{{root_dir}}" && \
    GH_FIX="{{fix}}" tests/test-lint.sh || \
      echo "Run 'just lint-fix' to fix the files."

# Lint everything and try to fix it.
[group("lint")]
lint-fix: (lint "true")

# Run all unittests.
[group("unitest")]
unittests:
  cd "{{root_dir}}" && \
    tests/test-unittests.sh

[group("release")]
release-test *args:
  cd "{{root_dir}}/githooks" && \
    GORELEASER_CURRENT_TAG=v9.9.9 \
    goreleaser release --snapshot --clean --skip=sign --skip=publish --skip=validate "$@"

[group("release")]
release version update-info="":
  cd "{{root_dir}}" && \
    scripts/release.sh "{{version}}" "{{update-info}}"

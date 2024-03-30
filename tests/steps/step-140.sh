#!/usr/bin/env bash
# Test:
#   Run shared hooks with images.yaml and staged files as file
set -e
set -u

TEST_DIR=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC1091
. "$TEST_DIR/general.sh"

steps/step-134.sh --export-staged-files-as-file

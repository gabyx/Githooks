#!/bin/sh
# Base Git hook template from https://github.com/gabyx/githooks
#
# It allows you to have a .githooks folder per-project that contains
# its hooks to execute on various Git triggers.

GITHOOKS_RUNNER=$(command -v "githooks-runner" 2>/dev/null || git config githooks.runner)

# shellcheck disable=SC2181
if [ ! -x "$GITHOOKS_RUNNER" ]; then
    echo "! Either 'githooks-runner' must be in in your path or" >&2
    echo "  Git config value in 'githooks.runner' must point to an " >&2
    echo "  executable. The value:" >&2
    echo "   '$GITHOOKS_RUNNER" >&2
    echo "  is not existing or is not executable!" >&2
    echo "  Please run the Githooks install script again to fix it." >&2
    exit 1
fi

exec "$GITHOOKS_RUNNER" "$0" "$@"

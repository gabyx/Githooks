#!/bin/sh
# Base Git hook template from https://github.com/gabyx/githooks
#
# It allows you to have a .githooks folder per-project that contains
# its hooks to execute on various Git triggers.

GITHOOKS_RUNNER=$(command -v "githooks-runner" >/dev/null 2>&1)

# shellcheck disable=SC2181
if [ $? -ne 0 ]; then
    GITHOOKS_RUNNER=$(git config githooks.runner)

    if [ ! -x "$GITHOOKS_RUNNER" ]; then
        echo "! Executable 'githooks-runner' is not in your path." >&2
        echo "! Also the optional Git config value in 'githooks.runner' points to " >&2
        echo "   \`$GITHOOKS_RUNNER\`" >&2
        echo " which is not existing or it is not executable!" >&2
        echo " Please run the Githooks install script again to fix it." >&2
        exit 1
    fi
fi

exec "$GITHOOKS_RUNNER" "$0" "$@"

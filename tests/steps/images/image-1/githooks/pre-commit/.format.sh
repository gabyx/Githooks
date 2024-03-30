#!/bin/bash
set -u
set -e

echo "Arguments given:" "$@"

if [ "${1:-}" != "--banana" ]; then
    echo "! First argument is not --banana"
    exit 1
fi

function do_stuff() {
    local file="$1"
    if [ ! -f "$file" ]; then
        echo "File '$file' does not exist!"
        exit 1
    fi
    echo "Formatting file '$file'..."
    echo "Formatted by containerized hook" >>"$file"
}

# Make sure we really do the right thing for the test.
if [ ! -f .githooks-test-export-staged-files ]; then
    [ -z "${STAGED_FILES_FILE:-}" ] || die "STAGED_FILES_FILE must not be defined."

    for file in $STAGED_FILES; do
        do_stuff "$file"
    done
else
    [ -z "${STAGED_FILES:-}" ] || die "STAGED_FILES must not be defined."

    while read -rd $'\0' file; do
        do_stuff "$file"
    done <"$STAGED_FILES_FILE"
fi

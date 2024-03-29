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

if [ -n "${STAGED_FILES:-}" ]; then
    for file in $STAGED_FILES; do
        do_stuff "$file"
    done
elif [ -n "${STAGED_FILES_FILE:-}" ]; then
    while read -rd $'\\0' file; do
        do_stuff "$file"
    done <"$STAGED_FILES_FILE"
fi

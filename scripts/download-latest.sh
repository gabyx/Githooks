#!/usr/bin/env bash

set -euo pipefail

ORG=gabyx
REPO=githooks

if ! command -v curl 2>/dev/null; then
    echo "Required tool curl is not installed."
    exit 1
fi

if ! command -v jq 2>/dev/null; then
    echo -e 'Required tool \033[1mjq\033[0m is not installed.'
    exit 1
fi

if [[ $# -eq 0 ]]; then
    # Find the latest version using the GitHub API
    SELECTED_TAG="$(curl --silent --location https://api.github.com/repos/${ORG}/${REPO}/releases | jq --raw-output 'map(select((.assets | length) > 0)) | .[0].tag_name')"
else
    # Use provided argument as tag to download
    SELECTED_TAG="$1"
fi

# Find a suitable install location
for CANDIDATE in "$HOME/bin" "/usr/local/bin" "/usr/bin"; do
    if [[ -w $CANDIDATE ]] && grep -q "$CANDIDATE" <<<"$PATH"; then
        TARGET_DIR="$CANDIDATE"
        break
    fi
done

# Bail out in case no suitable location could be found
if [[ -z ${TARGET_DIR:-} ]]; then
    echo -e "Unable to determine a writable install location. Make sure that you have write access to either \\033[1m/usr/local/bin\\033[0m or \\033[1m${HOME}/bin\\033[0m and that is in your PATH."
    exit 1
fi

SYSTEM_UNAME="$(uname | tr '[:upper:]' '[:lower:]')"
SYSTEM_ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')"

# Download and install
DOWNLOAD_URI="$(curl --silent --location "https://api.github.com/repos/${ORG}/${REPO}/releases/tags/${SELECTED_TAG}" | jq --raw-output ".assets[] | select( (.name | contains(\"${SYSTEM_UNAME}\")) and (.name | contains(\"${SYSTEM_ARCH}\")) ) | .browser_download_url")"
if [[ -z ${DOWNLOAD_URI} ]]; then
    echo -e "Unsupported operating system \\033[1m${SYSTEM_UNAME}\\033[0m or machine type \\033[1m${SYSTEM_ARCH}\\033[0m: Please check \\033[4;94mhttps://github.com/${ORG}/${REPO}/releases\\033[0m manually."
    exit 1
fi

echo -e "Downloading \\033[4;94m${DOWNLOAD_URI}\\033[0m to install \\033[1m${TARGET_DIR}/${REPO}\\033[0m"
case "${DOWNLOAD_URI}" in
*tar.gz)
    curl --progress-bar --location "${DOWNLOAD_URI}" | tar -xzf - -C "${TARGET_DIR}" "${REPO}"
    ;;

*)
    if curl --progress-bar --location "${DOWNLOAD_URI}" --output "${TARGET_DIR}/${REPO}"; then
        chmod a+rx "${TARGET_DIR}/${REPO}"
    fi
    ;;
esac

echo -e "\\nSuccessfully installed \\033[1m${TARGET_DIR}/${REPO}\\033[0m\\n"

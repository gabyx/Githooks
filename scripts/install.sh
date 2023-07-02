#!/usr/bin/env bash

set -euo pipefail

function checkTool() {
    if ! command -v "$1" &>/dev/null; then
        echo "!! Required tool '$1' is not installed."
        exit 1
    fi
}

checkTool "jq"
checkTool "curl"
checkTool "sha256sum"
checkTool "tar"
checkTool "unzip"
checkTool "uname"

org=gabyx
repo=githooks

unInstall="false"
installerArgs=()
versionTag=""

function printHelp() {
    echo -e "Usage: install.sh [options...] [-- <installer-args>...]\n\n" \
        "Options:\n" \
        "  --version <version> : The version to download (if not latest)\n" \
        "                        and install.\n" \
        "  --uninstall         : Uninstall Githooks. Uses always the latest uninstaller.\n" \
        "all other arguments are forwarded to the installer."
}

function parseArgs() {
    local toInstaller=false
    local prev=""

    for p in "$@"; do
        if [ "$toInstaller" = "true" ]; then
            installerArgs+=("$p")
        elif [ "$p" = "--help" ] || [ "$p" = "-h" ]; then
            printHelp
            exit 0
        elif [ "$p" = "--version" ]; then
            true
        elif [ "$p" = "--uninstall" ]; then
            unInstall="true"
        elif [ "$prev" = "--version" ]; then
            versionTag="v$p"

        elif [ "$p" = "--" ]; then
            toInstaller="true"
        else
            echo "! Unknown argument '$p'." >&2
            return 1
        fi

        prev="$p"
    done
}

parseArgs "$@"

if [ "$versionTag" = "" ] || [ "$unInstall" = "true" ]; then
    # Find the latest version using the GitHub API
    response=$(curl --silent --location "https://api.github.com/repos/$org/$repo/releases") || {
        echo "Could not get releases info from github.com"
        exit 1
    }

    versionTag="$(echo "$response" |
        jq --raw-output 'map(select((.assets | length) > 0)) | .[0].tag_name')"
fi

if ! echo "$versionTag" | grep -E "^v3"; then
    echo "!! Can only bootstrap version tags >= 'v3.0.0' with this script. Got tag '$versionTag'."
    # Older versions run the installer and directly update which
    # is a bit odd and setting up an older version is cumbersome.
    # exit 1
fi

systemName="$(uname | tr '[:upper:]' '[:lower:]')"
systemArch="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')"

# Download and install
response=$(curl --silent --location "https://api.github.com/repos/$org/$repo/releases/tags/$versionTag") || {
    echo "Could not get releases from github.com."
    exit 1
}

checksumFileURL=$(echo "$response" | jq --raw-output ".assets[] | select( .name == \"githooks.checksums\") | .browser_download_url")

url=$(echo "$response" |
    jq --raw-output ".assets[] | select( (.name | contains(\"$systemName\")) and (.name | contains(\"$systemArch\")) ) | .browser_download_url") || {
    echo "Could not get assets from tag '$versionTag'."
    exit 1
}

if [ -z "$url" ]; then
    echo -e "!! Unsupported operating system '$systemName' or \n" \
        "machine type '$systemArch': \n" \
        "Please check 'https://github.com/$org/${repo}/releases' manually."

    exit 1
fi

tempDir="$(mktemp -d)"

function cleanUp() {
    rm -rf "$tempDir" &>/dev/null || true
}
trap cleanUp EXIT

githooks="$tempDir/githooks"
mkdir -p "$githooks"

cliExe="cli"
if [ "$systemName" = "windows" ]; then
    cliExe="$cliExe.exe"
fi

cd "$tempDir"

echo -e "Downloading '$checksumFileURL'..."
checksums=$(curl --progress-bar --location "$checksumFileURL")

echo -e "Downloading '$url'..."
curl --progress-bar --location "$url" -o githooks.install

checksum=$(sha256sum "githooks.install" | cut -d ' ' -f 1)
if ! echo "$checksums" | grep -q "$checksum"; then
    echo "!! Checksum sha265 '$checksum' could not be verified in 'githooks.checksums' file."
    echo "$checksums"

    exit 1
else
    echo -e "\n=============================\nChecksums verified!\n=============================\n"
fi

case "$url" in
*.tar.gz)
    tar -C "$githooks" -xzf "githooks.install" >/dev/null
    ;;

*.zip)
    unzip -d "$githooks" "$githooks.install" >/dev/null
    ;;
esac

if [ "$unInstall" = "true" ]; then
    "githooks/$cliExe" uninstaller "${installerArgs[@]}"
else
    "githooks/$cliExe" installer "${installerArgs[@]}"
fi

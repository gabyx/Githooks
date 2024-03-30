#!/usr/bin/env bash
set -euo pipefail

function die() {
    echo -e "! ERROR:" "$@" >&2
    exit 1
}

function check_tool() {
    if ! command -v "$1" &>/dev/null; then
        echo "!! Required tool '$1' is not installed."
        exit 1
    fi
}

function check_bash() {
    if ! declare -n _DUMMY &>/dev/null; then
        die "You need bash at least 4.3 to run this script."
    fi
}

function get_platform_o_s() {
    local -n _platformOS="$1"
    local platformOSDist=""
    local platformOSVersion=""

    if [[ "$OSTYPE" == "linux"* ]]; then
        _platformOS="linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        _platformOS="darwin"
    elif [[ "$OSTYPE" == "freebsd"* ]]; then
        _platformOS="freebsd"
    else
        # Resort to `uname` for windows stuff.
        local name
        name=$(uname -a)
        case "$name" in
        CYGWIN*) _platformOS="windows" && platformOSDist="cygwin" ;;
        MINGW*) _platformOS="windows" && platformOSDist="mingw" ;;
        *Msys) _platformOS="windows" && platformOSDist="msys" ;;
        *) die "Platform: '$name' not supported." ;;
        esac
    fi

    if [ "$_platformOS" = "linux" ]; then

        if [ "$(lsb_release -si 2>/dev/null)" = "Ubuntu" ]; then
            platformOSDist="ubuntu"
            platformOSVersion=$(grep -m 1 "VERSION_CODENAME=" "/etc/os-release" |
                sed -E "s|.*=[\"']?(.*)[\"']?|\1|")
        elif grep -qE 'ID="?debian' "/etc/os-release"; then
            platformOSDist="debian"
            platformOSVersion=$(grep -m 1 "VERSION_CODENAME=" "/etc/os-release" |
                sed -E "s|.*=[\"']?(.*)[\"']?|\1|")
        elif grep -qE 'ID="?alpine' "/etc/os-release"; then
            platformOSDist="alpine"
            platformOSVersion=$(grep -m 1 "VERSION_ID=" "/etc/os-release" |
                sed -E 's|.*="?([0-9]+\.[0-9]+).*|\1|')
        elif grep -qE 'ID="?nixos' "/etc/os-release"; then
            platformOSDist="nixos"
            platformOSVersion=$(grep -m 1 "VERSION_ID=" "/etc/os-release" |
                sed -E 's|.*="?([0-9]+\.[0-9]+).*|\1|')
        elif grep -qE 'ID="?rhel' "/etc/os-release"; then
            platformOSDist="redhat"
            platformOSVersion=$(grep -m 1 "VERSION_ID=" "/etc/os-release" |
                sed -E 's|.*="?([0-9]+\.[0-9]+).*|\1|')
        elif grep -qE 'ID="?opensuse' "/etc/os-release"; then
            platformOSDist="opensuse"
            platformOSVersion=$(grep -m 1 "VERSION_ID=" "/etc/os-release" |
                sed -E 's|.*="?([0-9]+\.[0-9]+).*|\1|')
        else
            die "Linux Distro '/etc/os-release' not supported currently:" \
                "$(cat /etc/os-release 2>/dev/null)"
        fi

        # Remove whitespaces
        platformOSDist="${platformOSDist// /}"
        platformOSVersion="${platformOSVersion// /}"

    elif [ "$_platformOS" = "darwin" ]; then

        platformOSDist=$(sw_vers | grep -m 1 'ProductName' | sed -E 's/.*:\s+(.*)/\1/')
        platformOSVersion=$(sw_vers | grep -m 1 'ProductVersion' | sed -E 's/.*([0-9]+\.[0-9]+\.[0-9]+)/\1/g')
        # Remove whitespaces
        platformOSDist="${platformOSDist// /}"
        platformOSVersion="${platformOSVersion// /}"

    elif [ "$_platformOS" = "windows" ]; then
        platformOSVersion=$(systeminfo | grep -m 1 'OS Version:' | sed -E "s/.*:\s+([0-9]+\.[0-9]+\.[0-9]+).*/\1/")
        platformOSVersion="${platformOSVersion// /}"
    else
        die "Platform '$_platformOS' not supported currently."
    fi

    if [ -n "${2:-}" ]; then
        local -n _platformOSDist="$2"
        _platformOSDist="$platformOSDist"
    fi

    if [ -n "${3:-}" ]; then
        local -n _platformOSVersion="$3"
        _platformOSVersion="$platformOSVersion"
    fi

    return 0
}

function get_platform_arch() {
    local -n _arch="$1"

    _arch=""
    if uname -m | grep -q "x86_64" &>/dev/null; then
        _arch="amd64"
        return 0
    elif uname -m | grep -q -E "aarch64|arm64" &>/dev/null; then
        _arch="arm64"
        return 0
    elif uname -a | grep -q -E "aarch64|arm64" &>/dev/null; then
        _arch="arm64"
    else
        die "Architecture: '$(uname -m)' not supported."
    fi
}

check_bash
check_tool "grep"
check_tool "jq"
check_tool "curl"
check_tool "sha256sum"
check_tool "tar"
check_tool "unzip"
check_tool "uname"

org=gabyx
repo=githooks

unInstall="false"
installerArgs=()
versionTag=""

# Compare a and b as version strings. Rules:
# $1-a $2-op $3-b
# R1: a and b : dot-separated sequence of items. Items are numeric. The last item can optionally end with letters, i.e., 2.5 or 2.5a.
# R2: Zeros are automatically inserted to compare the same number of items, i.e., 1.0 < 1.0.1 means 1.0.0 < 1.0.1 => yes.
# R3: op can be '=' '==' '!=' '<' '<=' '>' '>=' (lexicographic).
# R4: Unrestricted number of digits of any item, i.e., 3.0003 > 3.0000004.
# R5: Unrestricted number of items.
function version_compare() {
    local a=$1 op=$2 b=$3 al=${1##*.} bl=${3##*.}
    while [[ $al =~ ^[[:digit:]] ]]; do al=${al:1}; done
    while [[ $bl =~ ^[[:digit:]] ]]; do bl=${bl:1}; done
    local ai=${a%"$al"} bi=${b%"$bl"}

    local ap=${ai//[[:digit:]]/} bp=${bi//[[:digit:]]/}
    ap=${ap//./.0} bp=${bp//./.0}

    local w=1 fmt=$a.$b x IFS=.
    for x in $fmt; do [ ${#x} -gt "$w" ] && w=${#x}; done
    fmt=${*//[^.]/}
    fmt=${fmt//./%${w}s}
    # shellcheck disable=SC2086,SC2059
    printf -v a "$fmt" $ai$bp
    printf -v a "%s-%${w}s" "$a" "$al"
    # shellcheck disable=SC2086,SC2059
    printf -v b "$fmt" $bi$ap
    printf -v b "%s-%${w}s" "$b" "$bl"

    # shellcheck disable=SC1072
    case $op in
    '<=' | '>=') test "$a" "${op:0:1}" "$b" || [ "$a" = "$b" ] ;;
    *) test "$a" "$op" "$b" ;;
    esac
}

function print_help() {
    echo -e "Usage: install.sh [options...] [-- <installer-args>...]\n\n" \
        "Options:\n" \
        "  --version <version> : The version to download (if not latest)\n" \
        "                        and install.\n" \
        "  --uninstall         : Uninstall Githooks. Uses always the latest uninstaller.\n" \
        "all other arguments are forwarded to the installer."
}

function parse_args() {
    local toInstaller=false
    local prev=""

    for p in "$@"; do
        if [ "$toInstaller" = "true" ]; then
            installerArgs+=("$p")
        elif [ "$p" = "--help" ] || [ "$p" = "-h" ]; then
            print_help
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

parse_args "$@"

if [ "$versionTag" = "" ] || [ "$unInstall" = "true" ]; then
    # Find the latest version using the GitHub API
    response=$(curl --silent --location "https://api.github.com/repos/$org/$repo/releases") || {
        echo "Could not get releases info from github.com"
        exit 1
    }

    versionTag="$(echo "$response" |
        jq --raw-output 'map(select((.assets | length) > 0)) | .[0].tag_name')"
fi

if ! version_compare "${versionTag##v}" ">=" "2.3.4"; then
    echo "!! Can only bootstrap version tags >= 'v2.3.4' with this script. Got tag '$versionTag'."
    exit 1
fi

os=""
arch=""

get_platform_o_s os
get_platform_arch arch

# The download used `macos` for `darwin` platform.
if [ "$os" = "darwin" ]; then
    os="macos"
fi

# Download and install
response=$(curl --silent --location "https://api.github.com/repos/$org/$repo/releases/tags/$versionTag") || {
    echo "Could not get releases from github.com."
    exit 1
}

checksumFileURL=$(echo "$response" | jq --raw-output ".assets[] | select( .name == \"githooks.checksums\") | .browser_download_url")

url=$(echo "$response" |
    jq --raw-output ".assets[] | select( (.name | contains(\"$os\")) and (.name | contains(\"$arch\")) ) | .browser_download_url") || {
    echo "Could not get assets from tag '$versionTag'."
    exit 1
}

if [ -z "$url" ]; then
    echo -e "!! Unsupported operating system '$os' or \n" \
        "machine type '$arch': \n" \
        "Please check 'https://github.com/$org/${repo}/releases' manually."

    exit 1
fi

tempDir="$(mktemp -d)"

function clean_up() {
    rm -rf "$tempDir" &>/dev/null || true
}
trap clean_up EXIT

githooks="$tempDir/githooks"
mkdir -p "$githooks"

cliExe="cli"
if [ "$os" = "windows" ]; then
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

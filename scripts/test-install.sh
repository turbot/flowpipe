#!/bin/sh
# Test installation script for bash. Mirrors install.sh script in the root directory. Any changes must be reflected in both scripts

# Check if exactly one argument is given
if [ "$#" -ne 1 ]; then
    echo "Error: This script requires exactly one argument which is the version of Flowpipe to install. For example: sudo ./test-install.sh v0.2.0-rc.2"
    exit 1
fi


# Function to check if a command exists
command_exists() {
    type "$1" &> /dev/null
}

set -e

FLOWPIPE_TEST_VERSION=$1
flowpipe_uri="https://github.com/turbot/flowpipe/releases/download/${FLOWPIPE_TEST_VERSION}/flowpipe.linux.arm64.tar.gz"

bin_dir="/usr/local/bin"
exe="$bin_dir/flowpipe"

test -z "$tmp_dir" && tmp_dir="$(mktemp -d)"
mkdir -p "${tmp_dir}"
tmp_dir="${tmp_dir%/}"

echo "Created temporary directory at $tmp_dir. Changing to $tmp_dir"
cd "$tmp_dir"

# set a trap for a clean exit - even in failures
trap 'rm -rf $tmp_dir' EXIT

case $(uname -s) in
	"Darwin") zip_location="$tmp_dir/flowpipe.tar.gz" ;;
	"Linux") zip_location="$tmp_dir/flowpipe.tar.gz" ;;
	*) echo "Error: flowpipe is not supported on '$(uname -s)' yet." 1>&2;exit 1 ;;
esac

echo "Downloading from $flowpipe_uri"
if command -v wget >/dev/null; then
	# because --show-progress was introduced in 1.16.
	wget --help | grep -q '\--showprogress' && _FORCE_PROGRESS_BAR="--no-verbose --show-progress" || _FORCE_PROGRESS_BAR=""
	# prefer an IPv4 connection, since github.com does not handle IPv6 connections properly.
	# Refer: https://github.com/turbot/steampipe/issues/861
	if ! wget --prefer-family=IPv4 --progress=bar:force:noscroll $_FORCE_PROGRESS_BAR -O "$zip_location" "$flowpipe_uri"; then
        echo "Could not find version $1"
        exit 1
    fi
elif command -v curl >/dev/null; then
	# curl uses HappyEyeball for connections, therefore, no preference is required
    if ! curl --fail --location --progress-bar --output "$zip_location" "$flowpipe_uri"; then
        echo "Could not find version $1"
        exit 1
    fi
else
    echo "Unable to find wget or curl. Cannot download."
    exit 1
fi

echo $zip_location
echo $tmp_dir

echo "Deflating downloaded archive"
tar -xvf "$zip_location" -C "$tmp_dir"

echo "Installing"
install -d "$bin_dir"
install "$tmp_dir/flowpipe" "$bin_dir"

echo "Applying necessary permissions"
chmod +x $exe

echo "Removing downloaded archive"
rm "$zip_location"

echo "Flowpipe was installed successfully to $exe"

if ! command -v $bin_dir/flowpipe >/dev/null; then
	echo "Flowpipe was installed, but could not be executed. Are you sure '$bin_dir/flowpipe' has the necessary permissions?"
	exit 1
fi


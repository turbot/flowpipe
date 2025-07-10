#!/bin/sh
# This is a a script to install dependencies/packages, create user, and assign necessary permissions in the ubuntu 24 container.
# Used in release smoke tests.

set -e

# update apt and install required packages
apt-get update
apt-get install -y tar ca-certificates jq curl git

# Extract the flowpipe binary
if [ "$(uname -m)" = "aarch64" ]; then
  tar -xzf /artifacts/linux-arm.tar.gz -C /usr/local/bin
else
  tar -xzf /artifacts/linux.tar.gz -C /usr/local/bin
fi

# Make the binary executable
chmod +x /usr/local/bin/flowpipe

# Make the scripts executable
chmod +x /scripts/smoke_test.sh

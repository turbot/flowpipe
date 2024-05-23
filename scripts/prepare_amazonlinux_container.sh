#!/bin/sh
# This is a a script to install dependencies/packages, create user, and assign necessary permissions in the amazonlinux 2023 container.
# Used in release smoke tests.

set -e

# update yum and install required packages
yum install -y shadow-utils tar gzip ca-certificates jq curl git --allowerasing

# Extract the powerpipe binary
tar -xzf /artifacts/linux.tar.gz -C /usr/local/bin

# Make the binary executable
chmod +x /usr/local/bin/flowpipe
 
# Make the scripts executable
chmod +x /scripts/smoke_test.sh

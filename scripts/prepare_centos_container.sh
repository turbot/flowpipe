#!/bin/sh
# This is a a script to install dependencies/packages, create user, and assign necessary permissions in the centos 9 container.
# Used in release smoke tests. 

set -e

# update yum and install required packages
yum install -y epel-release tar ca-certificates jq curl git --allowerasing

# Extract the flowpipe binary
tar -xzf  /artifacts/linux.tar.gz -C /usr/local/bin

# Make the binary executable
chmod +x /usr/local/bin/flowpipe

# Make the scripts executable
chmod +x /scripts/smoke_test.sh
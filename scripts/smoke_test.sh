#!/bin/sh
# This is a script with set of commands to smoke test a flowpipe build.
# The plan is to gradually add more tests to this script.

set -e

/usr/local/bin/flowpipe --version # check version
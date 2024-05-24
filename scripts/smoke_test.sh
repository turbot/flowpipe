#!/bin/sh
# This is a script with set of commands to smoke test a flowpipe build.
# The plan is to gradually add more tests to this script.

set -e

/usr/local/bin/flowpipe --version # check version

# Test the flowpipe repository to run test pipelines
git clone https://github.com/turbot/flowpipe.git
cd flowpipe

# List pipelines
/usr/local/bin/flowpipe pipeline list --mod-location internal/tests/test_pipelines/

# Run the test pipeline and capture output
output=$(/usr/local/bin/flowpipe pipeline run local.pipeline.simple_with_trigger --mod-location internal/tests/test_pipelines/ 2>&1)

# Print the output for debugging
echo "$output"

# Check for "Error" in the output
if echo "$output" | grep -q "Error"; then
  echo "Error found in pipeline run output"
  exit 1
fi

echo "Pipeline run completed successfully"
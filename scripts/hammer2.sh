#!/bin/bash

# Set the endpoint URL and data for the curl command
URL="https://localhost:7103/api/v0/play"
DATA='{"key":"key_placeholder","value":"value_placeholder"}'

# Parse the command line arguments
if [ $# -ne 2 ]
then
  echo "Usage: $0 <total_keys> <max_parallel>"
  exit 1
fi

TOTAL_KEYS="$1"
MAX_PARALLEL="$2"

# Generate random keys and store them in an array
for ((i=0; i<TOTAL_KEYS; i++))
do
  KEYS[$i]="$RANDOM"
done

# Define a function to make a single curl request with a given key
make_curl_request() {
  # Get the current key from the function argument
  local KEY="$1"

  # Replace the key_placeholder and value_placeholder in the data string with the current key and value
  local DATA="${DATA/key_placeholder/$KEY}"
  local DATA="${DATA/value_placeholder/$i}"

  # Call curl with the updated data string
  curl --insecure -X POST "$URL" --data "$DATA"
}

# Loop through the keys and make a curl request for each one in parallel
for ((i=0; i<TOTAL_KEYS; i++))
do
  # Get the current key from the array
  KEY="${KEYS[$i]}"

  # Call the make_curl_request function in the background with the current key
  make_curl_request "$KEY" &

  # If the maximum number of background jobs are running, wait for one to finish before starting a new one
  if (( $(jobs -r | wc -l) >= MAX_PARALLEL ))
  then
    # Wait for any background job to finish before continuing
    while (( $(jobs -r | wc -l) >= MAX_PARALLEL ))
    do
      sleep 0.1
    done
  fi
done

# Wait for all background jobs to finish before exiting the script
wait


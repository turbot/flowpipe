#!/bin/bash

# Set the endpoint URL and data for the curl command
URL="https://localhost:7103/api/v0/play"
DATA='{"key":"key_placeholder","value":"value_placeholder"}'

# Loop 1000 times and call curl with a different key each time
for ((i=1; i<=1000; i++))
do
  # Generate a random key using the current timestamp and the loop counter
  KEY="$(date +%s)_$i"

  # Replace the key_placeholder and value_placeholder in the data string with the current key and value
  DATA="${DATA/key_placeholder/$KEY}"
  DATA="${DATA/value_placeholder/$i}"

  # Call curl with the updated data string
  curl --insecure -X POST "$URL" --data "$DATA"

  # Reset the data string to its original value for the next iteration
  DATA='{"key":"key_placeholder","value":"value_placeholder"}'
done


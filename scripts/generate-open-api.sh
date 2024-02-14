#!/bin/bash

OUTPUT_DIR=internal/service/api/docs
swag init -o ${OUTPUT_DIR} -g internal/service/api/index.go --parseDependency --exclude ./internal/function

# TODO:: sed is able to generate the temp file but somehow the permission is broken and it is not able to overwrite the file
# Using the belw workaround for now. Will take a deep dive later
rm -f /tmp/flowpipe_api_temp.json
sed -r 's/\/query\./\//g;' ${OUTPUT_DIR}/swagger.json > /tmp/flowpipe_api_temp.json; mv /tmp/flowpipe_api_temp.json ${OUTPUT_DIR}/swagger.json
sed -r 's/"query\./"/g;' ${OUTPUT_DIR}/swagger.json > /tmp/flowpipe_api_temp.json; mv /tmp/flowpipe_api_temp.json ${OUTPUT_DIR}/swagger.json
sed -r 's/\/fperr\./\//g;' ${OUTPUT_DIR}/swagger.json > /tmp/flowpipe_api_temp.json; mv /tmp/flowpipe_api_temp.json ${OUTPUT_DIR}/swagger.json
sed -r 's/"fperr\./"/g;' ${OUTPUT_DIR}/swagger.json > /tmp/flowpipe_api_temp.json; mv /tmp/flowpipe_api_temp.json ${OUTPUT_DIR}/swagger.json
sed -r 's/\/types\./\//g;' ${OUTPUT_DIR}/swagger.json > /tmp/flowpipe_api_temp.json; mv /tmp/flowpipe_api_temp.json ${OUTPUT_DIR}/swagger.json
sed -r 's/"types\./"/g;' ${OUTPUT_DIR}/swagger.json > /tmp/flowpipe_api_temp.json; mv /tmp/flowpipe_api_temp.json ${OUTPUT_DIR}/swagger.json
curl -X POST https://converter.swagger.io/api/convert -d @${OUTPUT_DIR}/swagger.json --header 'Content-Type: application/json' > ${OUTPUT_DIR}/openapi.json

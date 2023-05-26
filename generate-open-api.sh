#!/bin/bash

OUTPUT_DIR=service/api/docs
swag init -o ${OUTPUT_DIR} -g service/api/api.go
sed -r -i 's/\/query\./\//g;' ${OUTPUT_DIR}/swagger.json
sed -r -i 's/"query\./"/g;' ${OUTPUT_DIR}/swagger.json
sed -r -i 's/\/sperr\./\//g;' ${OUTPUT_DIR}/swagger.json
sed -r -i 's/"sperr\./"/g;' ${OUTPUT_DIR}/swagger.json
sed -r -i 's/\/types\./\//g;' ${OUTPUT_DIR}/swagger.json
sed -r -i 's/"types\./"/g;' ${OUTPUT_DIR}/swagger.json
curl -X POST https://converter.swagger.io/api/convert -d @${OUTPUT_DIR}/swagger.json --header 'Content-Type: application/json' > ${OUTPUT_DIR}/openapi.json

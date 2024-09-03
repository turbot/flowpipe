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

echo ""
echo "Calling swagger.io to convert swagger.json to openapi.json"
echo ""
curl -X POST https://converter.swagger.io/api/convert -d @${OUTPUT_DIR}/swagger.json --header 'Content-Type: application/json' > ${OUTPUT_DIR}/openapi.json

echo ""
echo "Post processing openapi.json"
echo ""

# Swaggo does not support AnyType which we need for the default and value fields of FpVariable. Otherwise they will be mapped to object and will be generated as map[string]interface{}
jq '.components.schemas.FpVariable.properties.value_default.type = "AnyType" | .components.schemas.FpVariable.properties.value.type = "AnyType" | .components.schemas.FpVariable.properties.type.type = "AnyType" | .components.schemas.FpPipelineParam.properties.default.type = "AnyType" | .components.schemas.FpPipelineParam.properties.type.type = "AnyType"'  ${OUTPUT_DIR}/openapi.json > /tmp/flowpipe_openapi_api_temp.json; mv /tmp/flowpipe_openapi_api_temp.json ${OUTPUT_DIR}/openapi.json

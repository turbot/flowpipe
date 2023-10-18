FROM alpine:latest

# Install AWS CLI
RUN apk --no-cache add aws-cli

COPY aws-credentials /root/.aws/credentials
COPY parameter.json /root/parameter.json
COPY parameter_tef.json /root/parameter_tef.json
COPY parameter_ted.json /root/parameter_ted.json
COPY parameter_ted_create.json /root/parameter_ted_create.json

# Run the AWS CLI command to test the configuration
# CMD ["aws", "servicecatalog", "provision-product", "--product-name", "Turbot Guardrails Enterprise Database", "--provisioning-artifact-name", "v1.37.0", "--provisioning-parameters", "file://root/parameter_ted_create.json", "--provisioned-product-name", "ted-turbie3", "--region", "${param.region}"]
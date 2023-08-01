FROM mcr.microsoft.com/devcontainers/go:0-1.20-bullseye

RUN apt update \
    && apt install default-jre default-jdk -y \
    && cd /tmp \
    && wget https://dlcdn.apache.org/maven/maven-3/3.9.3/binaries/apache-maven-3.9.3-bin.tar.gz \
    && tar xvfz apache-maven-3.9.3-bin.tar.gz \
    && mv apache-maven-3.9.3 /usr/local/bin/ \
    && echo "export PATH=/usr/local/bin/apache-maven-3.9.3/bin:$PATH" >> /home/vscode/.profile

# IMPORTANT: need to switch the user before running Go package install
# otherwise the first Go install will be as root and the /go/pkg directory will be owned by root
USER vscode

RUN go install github.com/swaggo/swag/cmd/swag@v1.8.12 \
    && go install github.com/goreleaser/goreleaser@v1.18.2 \
    && go install github.com/mailhog/MailHog@v1.0.1 \
    && cd ~ \
    && mkdir openapi-generator \
    && cd openapi-generator \
    && wget https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/6.6.0/openapi-generator-cli-6.6.0.jar -O openapi-generator-cli.jar


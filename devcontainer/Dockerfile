FROM mcr.microsoft.com/devcontainers/go:0-1.20-bullseye

RUN apt update \
    && apt install default-jre default-jdk -y \
    && cd /tmp \
    && wget https://dlcdn.apache.org/maven/maven-3/3.9.2/binaries/apache-maven-3.9.2-bin.tar.gz \
    && tar xvfz apache-maven-3.9.2-bin.tar.gz \
    && mv apache-maven-3.9.2 /usr/local/bin/ \
    && echo "export PATH=/usr/local/bin/apache-maven-3.9.2/bin:$PATH" >> /home/vscode/.profile

# IMPORTANT to switch the user before running go install
# otherwise the first go install will be as root and the /go/pkg directory will be owned by root!
USER vscode

RUN go install github.com/swaggo/swag/cmd/swag@v1.8.12

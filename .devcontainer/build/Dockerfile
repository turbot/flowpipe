FROM mcr.microsoft.com/devcontainers/go:1-1.22-bullseye

ARG DEVCONTAINER_VERSION

LABEL org.opencontainers.image.ref.name="flowpipe-devcontainer" \
      org.opencontainers.image.version="${DEVCONTAINER_VERSION}" \
      org.opencontainers.image.url="https://flowpipe.io" \
      org.opencontainers.image.authors="Turbot HQ, Inc" \
      org.opencontainers.image.source="https://github.com/turbot/flowpipe" \
      org.opencontainers.image.description="Development container for flowpipe"

# This RUN command performs multiple setup tasks for the development environment:
# 1. Updates the package lists for upgrades and new package installations.
# 2. Installs Java Development Kit (JDK), Java Runtime Environment (JRE), and databases (PostgreSQL, MariaDB, SQLite3) for various development needs.
# 3. Downloads and installs Apache Maven for Java project management and build.
# 4. Adds Maven to the PATH environment variable for easy command-line access.
# 5. Sets up Node.js from the Nodesource repository, enabling the use of modern JavaScript tooling.
# 6. Enables Yarn package manager via Corepack for managing JavaScript packages.
# 7. Configures the `vscode` user to execute sudo commands without a password for easier container management.
# 8. Modifies PostgreSQL's authentication method to md5 for increased security.
# 9. Cleans up the apt cache and temporary files to reduce the image size.
RUN apt update \
    && apt install -y default-jre default-jdk postgresql mariadb-server sqlite3 \
    && wget -q https://dlcdn.apache.org/maven/maven-3/3.9.4/binaries/apache-maven-3.9.4-bin.tar.gz -P /tmp \
    && tar -xzf /tmp/apache-maven-3.9.4-bin.tar.gz -C /usr/local/bin \
    && echo "export PATH=/usr/local/bin/apache-maven-3.9.4/bin:$PATH" >> /home/vscode/.profile \
    && curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && corepack enable \
    && echo 'vscode ALL=(ALL) NOPASSWD: ALL' >> /etc/sudoers \
    && sed -i 's/local   all             all                                     peer/local   all             all                                     md5/' /etc/postgresql/*/main/pg_hba.conf \
    && rm -rf /var/lib/apt/lists/*

# IMPORTANT: need to switch the user before running Go package install
# otherwise the first Go install will be as root and the /go/pkg directory will be owned by root
USER vscode

# This RUN command sets up essential Go packages and tools for API documentation, release management, and email testing, along with OpenAPI Generator and Yarn package manager:
# 1. Installs Swag for automatically generating RESTful API documentation with Swagger 2.0.
# 2. Installs GoReleaser, a tool to streamline the release process for Go applications, supporting various platforms and packaging methods.
# 3. Installs MailHog, an email testing tool for developers, allowing the interception and viewing of emails sent from the development environment.
# 4. Creates a directory for the OpenAPI Generator, which is used for generating API clients, server stubs, and API documentation from OpenAPI Specifications.
# 5. Downloads the OpenAPI Generator CLI jar file, enabling the generation of API clients and documentation from a command line interface.
# 6. Activates the Yarn package manager via Corepack, preparing the environment for JavaScript package management and development.
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.3 \
    && go install github.com/goreleaser/goreleaser@v2.0.1 \
    && go install github.com/mailhog/MailHog@v1.0.1 \
    && mkdir -p /home/vscode/openapi-generator \
    && wget https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/6.6.0/openapi-generator-cli-6.6.0.jar -O /home/vscode/openapi-generator/openapi-generator-cli.jar \
    && corepack prepare yarn@stable --activate


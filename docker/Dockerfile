# Use the slim version of Debian Bullseye as the base image
FROM debian:bullseye-slim

LABEL org.opencontainers.image.ref.name flowpipe \
      org.opencontainers.image.version ${TARGETVERSION} \
      org.opencontainers.image.url "https://flowpipe.io" \
      org.opencontainers.image.authors "Turbot HQ, Inc" \
      org.opencontainers.image.source "https://github.com/turbot/flowpipe" \
      org.opencontainers.image.description "Flowpipe container image ${TARGETVERSION}"

# Set environment variables
# USER_UID: Override the flowpipe UID (default: 7103)
# USER_GID: Override the flowpipe GID (default: 0)
# FLOWPIPE_LOG_LEVEL: Set logging level (default: "off")
# FLOWPIPE_UPDATE_CHECK: Disable auto-update checks (default: false)
# FLOWPIPE_TELEMETRY: Disable telemetry (default: none)
ENV USER_UID=7103 \
    USER_GID=0 \
    FLOWPIPE_LOG_LEVEL="off" \
    FLOWPIPE_UPDATE_CHECK=false \
    FLOWPIPE_TELEMETRY=none

# Declare build arguments
# TARGETVERSION: Specify the version to target
# TARGETARCH: Specify the architecture to target
ARG TARGETVERSION \
    TARGETARCH

# This RUN command performs several operations essential for setting up the container:
# 1. Determines the group name based on USER_GID and creates a system user 'flowpipe' with specified UID and no login shell. This user is for running flowpipe processes.
# 2. Updates the package lists to ensure we have access to the latest versions of packages and their dependencies.
# 3. Installs necessary packages for the setup:
#    - apt-transport-https: Allows the use of repositories accessed via the HTTPS protocol.
#    - ca-certificates: Ensures that SSL/TLS certificates can be verified.
#    - curl: Used for transferring data with URLs.
#    - gnupg: Provides encryption services and key management.
#    - lsb-release: Provides information about the Linux distribution.
#    - gosu: A simple tool used to step down from root and execute commands as the flowpipe user.
#    - wget: A network downloader.
# 4. Downloads and adds the Docker repository GPG key to ensure software authenticity.
# 5. Adds the Docker repository to apt sources list, enabling installation of Docker packages.
# 6. Performs another apt-get update to recognize the newly added Docker repository.
# 7. Installs the Docker CLI to enable container management within the container.
# 8. Creates a directory for flowpipe and downloads the flowpipe binary package.
# 9. Extracts the flowpipe package to /opt/flowpipe and moves the flowpipe executable to /usr/local/bin for easy execution.
# 10. Cleans up the apt cache and removes temporary files to reduce the image size.
# Note: Directory: `/opt/flowpipe` is where the contents of the tar file are extracted.
# Note: Directory: `/usr/local/bin` is where the flowpipe binary is moved to.
# Note: Packages apt-transport-https, ca-certificates, curl, gnupg, lsb-release, docker-ce-cli are required for Docker
# Note: Second apt-get cli is used to install docker-ce-cli
RUN group_name=$(getent group ${USER_GID} | cut -d: -f1) && \
    adduser --system --disabled-login --ingroup $group_name --gecos "flowpipe user" --shell /bin/false --uid $USER_UID flowpipe && \
    apt-get update && \
    apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release gosu wget && \
    curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list && \
    apt-get update -y && \
    apt-get install -y docker-ce-cli && \
    mkdir -p /opt/flowpipe && \
    wget -nv https://github.com/turbot/flowpipe/releases/download/${TARGETVERSION}/flowpipe.linux.${TARGETARCH}.tar.gz -O /tmp/flowpipe.linux.${TARGETARCH}.tar.gz && \
    tar xzf /tmp/flowpipe.linux.${TARGETARCH}.tar.gz -C /opt/flowpipe && \
    mv /opt/flowpipe/flowpipe /usr/local/bin/flowpipe && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/flowpipe.linux.${TARGETARCH}.tar.gz

# Creates the workspace directory with the correct ownership
USER flowpipe
WORKDIR /workspace
USER root

# Expose port 7103 for flowpipe
EXPOSE 7103

# Copy the entrypoint script into the image
COPY docker-entrypoint.sh /usr/local/bin

# Define the entrypoint and default command
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["flowpipe"]

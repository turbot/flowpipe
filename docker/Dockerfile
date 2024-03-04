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

# Install gosu to enable a smooth switch from the root user to a non-root user in the Docker container.
# Add a non-root user 'flowpipe' for security purposes,
# avoid running the container as root, update the package list,
# install necessary packages for adding Docker's repository, add Docker’s official GPG key,
# set up the Docker stable repository, update the package list again,
# install 'wget' for downloading flowpipe, 'docker-ce-cli' for docker commands,
# download the release as specified in TARGETVERSION and TARGETARCH,
# extract it, move it to the appropriate directory, and then clean up.
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

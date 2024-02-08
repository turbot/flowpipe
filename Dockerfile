# Use the slim version of Debian Bullseye as the base image
FROM debian:bullseye-slim

LABEL org.opencontainers.image.ref.name=flowpipe
LABEL org.opencontainers.image.version=${TARGETVERSION}
LABEL org.opencontainers.image.url="https://flowpipe.io"
LABEL org.opencontainers.image.authors="Turbot HQ, Inc"
LABEL org.opencontainers.image.source="https://github.com/turbot/flowpipe"

# Text only description, don't change to use special characters
LABEL org.opencontainers.image.description="Flowpipe container image"

# Define default environment variables to override the flowpipe UID and its GID
ENV USER_UID=7103
ENV USER_GID=0

# Define default environment variables to enable debugging logging
ENV FLOWPIPE_LOG_LEVEL="off"

# Declare build arguments for version and architecture
ARG TARGETVERSION
ARG TARGETARCH

# Install gosu to enable a smooth switch from the root user to a non-root user in the Docker container.
# Add a non-root user 'flowpipe' for security purposes,
# avoid running the container as root, update the package list,
# install necessary packages for adding Docker's repository, add Docker’s official GPG key,
# set up the Docker stable repository, update the package list again,
# install 'wget' for downloading flowpipe, 'docker-ce-cli' for docker commands,
# download the release as specified in TARGETVERSION and TARGETARCH,
# extract it, move it to the appropriate directory, and then clean up.
RUN group_name=$(getent group ${USER_GID} | cut -d: -f1) && \
    adduser --system --disabled-login --ingroup $group_name --gecos "flowpipe user" --shell /bin/false --uid $USER_UID flowpipe && \
    apt-get update && \
    apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release gosu  && \
    curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list && \
    apt-get update -y && \
    apt-get install -y docker-ce-cli wget && \
    mkdir -p /opt/flowpipe && \
    wget -nv https://github.com/turbot/flowpipe/releases/download/${TARGETVERSION}/flowpipe.linux.${TARGETARCH}.tar.gz -O /tmp/flowpipe.linux.${TARGETARCH}.tar.gz && \
    tar xzf /tmp/flowpipe.linux.${TARGETARCH}.tar.gz -C /opt/flowpipe && \
    mv /opt/flowpipe/flowpipe /usr/local/bin/flowpipe && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/flowpipe.linux.${TARGETARCH}.tar.gz

# Expose port 7103 for flowpipe
EXPOSE 7103

# Set environment variables to disable auto-update and telemetry for flowpipe
ENV FLOWPIPE_UPDATE_CHECK=false
ENV FLOWPIPE_TELEMETRY=none

# Copy the entrypoint script into the image
COPY docker-entrypoint.sh /usr/local/bin

# Define the entrypoint and default command
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["flowpipe"]

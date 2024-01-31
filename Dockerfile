# Use the slim version of Debian Bullseye as the base image
FROM debian:bullseye-slim

# Label the image with the maintainer's contact information
# TODO: Confirm the maintainer details with @cody
LABEL maintainer="Turbot Support <help@turbot.com>"

# Declare build arguments for version and architecture
ARG TARGETVERSION
ARG TARGETARCH

# Add a non-root user 'flowpipe' for security purposes,
# avoid running the container as root, update the package list,
# install necessary packages for adding Docker's repository, add Docker’s official GPG key,
# set up the Docker stable repository, update the package list again,
# install 'wget' for downloading flowpipe, 'docker-ce-cli' for docker commands,
# download the release as specified in TARGETVERSION and TARGETARCH,
# extract it, move it to the appropriate directory, and then clean up.
RUN adduser --system --disabled-login --ingroup 0 --gecos "flowpipe user" --shell /bin/false --uid 7103 flowpipe && \
    apt-get update && \
    apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release && \
    curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list && \
    apt-get update -y && \
    apt-get install -y docker-ce-cli wget && \
    wget -nv https://github.com/turbot/flowpipe/releases/download/${TARGETVERSION}/flowpipe.linux.${TARGETARCH}.tar.gz -O /tmp/flowpipe.linux.${TARGETARCH}.tar.gz && \
    tar xzf /tmp/flowpipe.linux.${TARGETARCH}.tar.gz -C /usr/local/bin && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/flowpipe.linux.${TARGETARCH}.tar.gz

# Switch to the non-root user 'flowpipe'
USER flowpipe:0

# Set the working directory inside the container to /workspace.
# This directory can be mounted from the host machine.
WORKDIR /workspace

# Expose port 7103 for flowpipe
EXPOSE 7103

# Set environment variables to disable auto-update and telemetry for flowpipe
ENV FLOWPIPE_UPDATE_CHECK=false
ENV FLOWPIPE_TELEMETRY=none

# Create the flowpipe config directory with the correct permissions
# TODO: Confirm if the configuration for flowpipe is still correct
RUN mkdir -p /home/flowpipe/.flowpipe/config && \
    chown -R flowpipe:0 /home/flowpipe/.flowpipe && \
    chown -R flowpipe:0 /home/flowpipe/.flowpipe/config

# Copy the entrypoint script into the image
COPY docker-entrypoint.sh /usr/local/bin

# Define the entrypoint and default command
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["flowpipe"]

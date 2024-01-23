FROM debian:bullseye-slim
LABEL maintainer="Turbot Support <help@turbot.com>"

ARG TARGETVERSION
ARG TARGETARCH

# add a non-root 'flowpipe' user
RUN adduser --system --disabled-login --ingroup 0 --gecos "flowpipe user" --shell /bin/false --uid 9193 flowpipe

# updates and installs - 'wget' for downloading flowpipe, 'less' for paging in 'flowpipe query' interactive mode
RUN apt-get update -y && apt-get install -y wget less && rm -rf /var/lib/apt/lists/*

# download the release as given in TARGETVERSION and TARGETARCH
RUN echo \
 && cd /tmp \
 && wget -nv https://github.com/turbot/flowpipe/releases/download/${TARGETVERSION}/flowpipe.linux.${TARGETARCH}.tar.gz \
 && tar xzf flowpipe.linux.${TARGETARCH}.tar.gz \
 && mv flowpipe /usr/local/bin/ \
 && rm -rf /tmp/flowpipe.linux.${TARGETARCH}.tar.gz

# Change user to non-root
USER flowpipe:0

# Use a constant workspace directory that can be mounted to
WORKDIR /workspace

# disable auto-update
ENV FLOWPIPE_UPDATE_CHECK=false

# disable telemetry
ENV FLOWPIPE_TELEMETRY=none

COPY docker-entrypoint.sh /usr/local/bin
ENTRYPOINT [ "docker-entrypoint.sh" ]
CMD [ "flowpipe"]

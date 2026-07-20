FROM library/debian:trixie-slim AS slim
RUN --mount=type=cache,target=/var/lib/apt/lists/ \
  --mount=type=cache,target=/var/cache/apt/archives/ <<EOF
set -e
apt-get update
apt-get full-upgrade
apt-get install -y ca-certificates  ca-certificates \
    fonts-liberation \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libatspi2.0-0 \
    libc6 \
    libcairo2 \
    libcups2 \
    libdbus-1-3 \
    libdrm2 \
    libexpat1 \
    libgbm1 \
    libglib2.0-0 \
    libgtk-3-0 \
    libnspr4 \
    libnss3 \
    libpango-1.0-0 \
    libx11-6 \
    libx11-xcb1 \
    libxcomposite1 \
    libxdamage1 \
    libxext6 \
    libxfixes3 \
    libxkbcommon0 \
    libxrandr2 \
    wget
apt-get autoclean -y
apt-get autopurge -y
rm -rf /var/lib/apt/lists/*
useradd -m north_outage
EOF
COPY north_outage /usr/bin/north_outage

USER north_outage

RUN /usr/bin/north_outage setup

WORKDIR /home/north_outage

ENTRYPOINT ["/usr/bin/north_outage" ]
CMD ["--config","/config.yaml"]

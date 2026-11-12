#!/bin/bash
set -eo pipefail

PORT=$(curl -sf -H "Metadata-Flavor: Google" \
  http://metadata.google.internal/computeMetadata/v1/instance/attributes/httpbin-port)

# Install Docker from the Debian repositories
apt-get update
apt-get install -y docker.io

# The startup script runs on every boot, so only create the container once.
# --restart always brings it back after a reboot.
if ! docker inspect httpbin >/dev/null 2>&1; then
  docker run -d --name httpbin --restart always -p "$PORT:8080" mccutchen/go-httpbin
fi

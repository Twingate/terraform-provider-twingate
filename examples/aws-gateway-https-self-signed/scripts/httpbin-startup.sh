#!/bin/bash
set -eo pipefail

# Retry because outbound access through the NAT gateway can lag behind boot.
APT_OPTS="-o Acquire::Retries=10"

# Install Docker from the Ubuntu repositories
apt-get $APT_OPTS update
apt-get $APT_OPTS install -y docker.io

# --restart always brings the container back after a reboot.
docker run -d --name httpbin --restart always -p "${httpbin_port}:8080" mccutchen/go-httpbin
